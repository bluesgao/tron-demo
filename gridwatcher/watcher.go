package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
)

// ---------- Fetch + filter ----------

type EventWatcher struct {
	restyClient *resty.Client
}

// FetchTransferEvents fetches Transfer events from TronGrid API
// Parameters:
//   - ctx: context for cancellation
//   - fingerprint: pagination token from previous response (empty for first page)
//   - onlyConfirmed: if true, only returns confirmed events
//   - limit: max number of events per page (default 20, max 200)
//   - minBlockTimestamp: minimum block timestamp in milliseconds (0 to get all history)
//   - maxBlockTimestamp: maximum block timestamp in milliseconds (0 for no limit)
//   - orderBy: sort order, "block_timestamp,desc" (newest first) or "block_timestamp,asc" (oldest first)
//
// API Reference: https://developers.tron.network/reference/get-events-by-contract-address
func (w *EventWatcher) FetchTransferEvents(ctx context.Context, fingerprint string, onlyConfirmed bool, limit int, minBlockTimestamp, maxBlockTimestamp int64, orderBy string) (*TronGridResp, error) {
	url := fmt.Sprintf(TronGridUrl, UsdtContract)

	req := w.restyClient.R().
		SetContext(ctx).
		SetQueryParam("event_name", EventNameTransfer). // Filter by event name (Transfer for TRC-20 transfers)
		SetQueryParam("limit", fmt.Sprintf("%d", limit)). // Max 200 per page
		SetQueryParam("only_confirmed", fmt.Sprintf("%t", onlyConfirmed)) // Only confirmed events

	// Pagination: use fingerprint from previous response to get next page
	if fingerprint != "" {
		req.SetQueryParam("fingerprint", fingerprint)
		log.Printf("[FetchTransferEvents] Using fingerprint: %s", fingerprint)
	}

	// Time window: min_block_timestamp (inclusive, milliseconds)
	// Only get events after this timestamp (0 = no lower limit, get all history)
	if minBlockTimestamp > 0 {
		req.SetQueryParam("min_block_timestamp", fmt.Sprintf("%d", minBlockTimestamp))
		log.Printf("[FetchTransferEvents] Using min_block_timestamp: %d", minBlockTimestamp)
	}

	// Time window: max_block_timestamp (inclusive, milliseconds)
	// Only get events before this timestamp (0 = no upper limit)
	if maxBlockTimestamp > 0 {
		req.SetQueryParam("max_block_timestamp", fmt.Sprintf("%d", maxBlockTimestamp))
		log.Printf("[FetchTransferEvents] Using max_block_timestamp: %d", maxBlockTimestamp)
	}

	// Sort order: "block_timestamp,desc" (newest first, default) or "block_timestamp,asc" (oldest first)
	if orderBy != "" {
		req.SetQueryParam("order_by", orderBy)
		log.Printf("[FetchTransferEvents] Using order_by: %s", orderBy)
	}

	log.Printf("[FetchTransferEvents] Requesting events: onlyConfirmed=%v, limit=%d, minBlockTimestamp=%d, maxBlockTimestamp=%d, orderBy=%s",
		onlyConfirmed, limit, minBlockTimestamp, maxBlockTimestamp, orderBy)

	var result TronGridResp
	resp, err := req.SetResult(&result).Get(url)

	if err != nil {
		log.Printf("[FetchTransferEvents] Request failed: %v", err)
		return nil, err
	}

	if !resp.IsSuccess() {
		log.Printf("[FetchTransferEvents] HTTP error: status=%d, body=%s", resp.StatusCode(), resp.String())
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode(), resp.String())
	}

	log.Printf("[FetchTransferEvents] Success: received %d events, fingerprint=%s", len(result.Data), result.Meta.Fingerprint)
	return &result, nil
}

func main() {
	log.Println("[main] Starting TronGrid event watcher")

	// 1) 你的关注地址池（示例：base58）
	watchBase58 := []string{
		"TNPdqto8HiuMzoG7Vv9wyyYhWzCojLeHAF", // Binance-Cold 4
	}

	log.Printf("[main] Monitoring %d addresses", len(watchBase58))

	// 2) 预处理成 hex set
	watchHex := make(map[string]struct{}, len(watchBase58))
	for _, a := range watchBase58 {
		h, err := TronBase58ToEvmHex(a)
		if err != nil {
			log.Fatalf("[main] Failed to convert address %s: %v", a, err)
		}
		watchHex[h] = struct{}{}
		log.Printf("[main] Watching address: base58=%s, hex=%s", a, h)
	}

	if TronGridAPIKey == "" {
		log.Fatal("[main] API Key is empty, please set a valid TronGrid API Key")
	}
	log.Println("[main] API Key configured")

	restyClient := resty.New().
		SetTimeout(DefaultTimeout).
		SetHeader(HeaderTronProAPIKey, TronGridAPIKey)

	watcher := &EventWatcher{
		restyClient: restyClient,
	}

	ctx := context.Background()
	// 监控从当前时间往前推指定时间窗口到当前时间的全部转账交易
	orderBy := "block_timestamp,desc" // 降序：最新的在前

	log.Printf("[main] Starting to monitor events from (current time - %v) to current time", LookbackWindow)

	for {
		// 外层循环：每次重新计算时间窗口
		// 固定时间窗口，避免在分页过程中时间窗口变化导致 fingerprint 失效
		currentTime := time.Now()
		startTime := currentTime.Add(-LookbackWindow)
		minBlockTimestamp := startTime.UnixMilli()
		maxBlockTimestamp := currentTime.UnixMilli()

		log.Printf("[main] Starting new time window: %s to %s", startTime.Format(time.RFC3339), currentTime.Format(time.RFC3339))

		// 内层循环：使用固定的时间窗口进行分页
		fingerprint := ""
		pageCount := 0
		for {
			pageCount++
			log.Printf("[main] Fetching page %d (time range: %s to %s)", pageCount,
				startTime.Format(time.RFC3339),
				currentTime.Format(time.RFC3339))

			r, err := watcher.FetchTransferEvents(ctx, fingerprint, true /*onlyConfirmed*/, DefaultLimit, minBlockTimestamp, maxBlockTimestamp, orderBy)
			if err != nil {
				log.Fatalf("[main] Failed to fetch events: %v", err)
			}

			log.Printf("[main] Processing %d events from page %d", len(r.Data), pageCount)

			matchedCount := 0
			// 处理事件
			for _, ev := range r.Data {
				// 只处理 Transfer（我们请求里已经指定 Transfer，这里再保险）
				if ev.EventName != EventNameTransfer {
					continue
				}

				to := ev.ToHex()
				from := ev.FromHex()

				// 只监控to地址的交易
				isToWatched := true
				//if _, ok := watchHex[to]; ok {
				//	isToWatched = true
				//}

				// 如果 to 是监控地址，则处理
				if isToWatched {
					matchedCount++
					// 幂等键：txid + event_index
					idempotencyKey := fmt.Sprintf("%s%s%d", ev.TransactionID, IdempotencyKeySeparator, ev.EventIndex)
					eventType := "DEPOSIT"
					log.Printf("[main] %s hit: to=%s from=%s value=%s tx=%s confirmed=%v key=%s",
						eventType, to, from, ev.ValueStr(), ev.TransactionID, !ev.Unconfirmed, idempotencyKey)

					// TODO: 入库（unique(idempotencyKey)）+ 触发入账流程
				}
			}

			if matchedCount > 0 {
				log.Printf("[main] Found %d matched events on page %d", matchedCount, pageCount)
			}

			// 如果到达数据末尾（没有更多数据），退出内层循环，重新计算时间窗口
			if r.Meta.Fingerprint == "" || len(r.Data) == 0 {
				log.Printf("[main] Reached end of current time window: fingerprint=%s, dataCount=%d", r.Meta.Fingerprint, len(r.Data))
				break
			}

			// 继续分页获取（使用固定的时间窗口）
			fingerprint = r.Meta.Fingerprint
			log.Printf("[main] Continuing to next page with fingerprint: %s", fingerprint)
		}

		// 完成当前时间窗口的所有分页后，等待一段时间，然后重新计算新的时间窗口
		log.Printf("[main] Completed time window, waiting %v before checking next time window", PollInterval)
		time.Sleep(PollInterval)
	}
}
