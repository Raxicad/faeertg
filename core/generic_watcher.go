package core

import (
	"context"
	"errors"
	"fmt"
	"go.elastic.co/apm"
	"strconv"
	"strings"
	"sync"
)

type executionResult struct {
	target     WatchTarget
	ctx        context.Context
	wg         *sync.WaitGroup
	currStatus ProductStatus
	prevStatus ProductStatus
	//product WebhookPayload
	//tx      *apm.Transaction
}

type genericWatcher struct {
	statusChangeHandler StatusChangeHandler
	ctx                 context.Context
	internalCtx         context.Context
	cancelFunc          context.CancelFunc
	dataChan            chan *executionResult
	fetcherFactory      ProductStatusFetcherFactory
	monitorSlug         string
	fetchUrl            string
	results             map[string]*executionResult
	mu                  *sync.Mutex
	statuses            map[string]ProductStatus
	targetsCount        int
	//status              ProductStatus
	//wg                  *sync.WaitGroup
}

//
//func (tw *genericWatcher) GetProductTitle() string {
//	return tw.productTitle
//}
//
//func (tw *genericWatcher) GetProductPic() string {
//	return tw.productPic
//}

func (tw *genericWatcher) Dispose() {
	//tw.wg.Done()
}

func (tw *genericWatcher) Spawn(instancesCount int) {
	//tw.wg.Add(1)
	inctx, cfn := context.WithCancel(context.Background())
	tw.internalCtx = inctx
	tw.cancelFunc = cfn
	tw.dataChan = make(chan *executionResult, instancesCount*tw.targetsCount)

	go func() {
		for {
			select {
			case <-tw.ctx.Done():
				tw.cancelFunc()
				continue
			case <-tw.internalCtx.Done():
				//tw.wg.Done()
				return
			case d := <-tw.dataChan:
				go func(data *executionResult) {
					tw.mu.Lock()
					tw.results[data.target.GetTargetId()] = data
					tw.mu.Unlock()
					err := tw.statusChangeHandler(data.ctx, data.target, data.prevStatus, data.currStatus)
					if err != nil {
						_ = apm.CaptureError(data.ctx, err)
					}
					//tw.status = data.status
					//data.tx.End()
					data.wg.Done()
				}(d)
			}
		}
	}()

	go tw.executeCheckIfAvailable(instancesCount)
}

func (tw *genericWatcher) executeCheckIfAvailable(instancesCount int) {
	var fetchers []ProductStatusFetcher
	for i := 0; i < instancesCount; i++ {
		fetchers = append(fetchers, tw.fetcherFactory.CreateFetcher(tw.fetchUrl))
	}

	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for true {
		tx := apm.DefaultTracer.StartTransaction("check_status", tw.monitorSlug)
		tracingCtx := apm.ContextWithTransaction(tw.internalCtx, tx)
		//wg.Add(instancesCount)
		//status := Unavailable
		responsesMap := map[string]ProductStatus{}
		wg.Add(tw.targetsCount * len(fetchers))
		for i, currentFetcher := range fetchers {
			go func(ix int, fetcher ProductStatusFetcher) {
				fetcherSpan, _ := apm.StartSpan(tracingCtx, fmt.Sprintf("check_worker_%d", ix), tw.monitorSlug)
				fetcherCtx := apm.ContextWithSpan(tracingCtx, fetcherSpan)
				defer fetcherSpan.End()
				result := fetcher.FetchStatus(fetcherCtx)
				fetcherSpan.SpanData.Context.SetHTTPRequest(result.Request)
				fetcherSpan.SpanData.Context.SetHTTPStatusCode(result.StatusCode)

				if result.Response != nil {
					for k, values := range result.Response.Header {
						fetcherSpan.SpanData.Context.SetLabel(k, strings.Join(values, "\n"))
					}
					//tx.TransactionData.Context.SetHTTPResponseHeaders(result.Response.Header)
				}

				if len(result.RawResponse) > 0 {
					fetcherSpan.SpanData.Context.SetLabel("response_payload", string(result.RawResponse))
				}

				if result.Error != nil {
					e := apm.CaptureError(fetcherCtx, result.Error)
					e.SetSpan(fetcherSpan)
					e.Send()
				}

				if tw.targetsCount != len(result.Targets) {
					err := errors.New("targets count returned by watchersFactory must match count of fetched targets")
					fatalErrTx := apm.DefaultTracer.StartTransaction("fatal_error", tw.monitorSlug)
					fatalErrTx.Context.SetHTTPStatusCode(result.StatusCode)
					fatalErrTx.Context.SetHTTPRequest(result.Request)
					if result.Response != nil {
						fatalErrTx.Context.SetHTTPResponseHeaders(result.Response.Header)
					}

					fatalErrTx.Context.SetLabel("expected_targets_count", tw.targetsCount)
					fatalErrTx.Context.SetLabel("actual_targets_count", len(result.Targets))
					fatalErrTx.Context.SetLabel("fetch_url", tw.fetchUrl)
					for ix, t := range result.Targets {
						fatalErrTx.Context.SetLabel("target_"+strconv.Itoa(ix), fmt.Sprintf("id='%s' available='%v' title='%s' pic='%s'", t.GetTargetId(), t.Available(), t.GetProductTitle(), t.GetProductPic()))
					}

					e := apm.CaptureError(apm.ContextWithTransaction(context.Background(), fatalErrTx), err)
					e.SetTransaction(fatalErrTx)
					e.Send()
					fatalErrTx.End()
					for i := 0; i < tw.targetsCount; i++ {
						wg.Done()
					}
					return
				}

				for _, target := range result.Targets {
					go func(t WatchTarget) {
						targetIterSpan, _ := apm.StartSpan(tracingCtx, fmt.Sprintf("target_%s", t.GetTargetId()), tw.monitorSlug)
						targetIterCtx := apm.ContextWithSpan(tracingCtx, targetIterSpan)
						defer targetIterSpan.End()
						iterationWg := &sync.WaitGroup{}
						iterationWg.Add(1)
						mu.Lock()
						currStatus, ok := responsesMap[t.GetTargetId()]
						if !ok {
							responsesMap[t.GetTargetId()] = Unavailable
							currStatus = Unavailable
						}

						if t.Available() {
							currStatus = Available
						}

						responsesMap[t.GetTargetId()] = currStatus
						prevStatus, _ := tw.statuses[t.GetTargetId()]
						mu.Unlock()

						tw.notify(t, targetIterCtx, iterationWg, prevStatus, currStatus)
						iterationWg.Wait()
						wg.Done()

						mu.Lock()
						tw.statuses[t.GetTargetId()] = currStatus
						mu.Unlock()
					}(target)
				}
				//
				//if result.Available {
				//	status = Available
				//}
				//
				//if len(result.ProductPic) > 0 {
				//	tw.productPic = result.ProductPic
				//}
				//
				//if len(result.ProductTitle) > 0 {
				//	tw.productTitle = result.ProductTitle
				//}
				//tw.notify(status, result.Payload, fetcherCtx, iterationWg)
			}(i, currentFetcher)
		}
		wg.Wait()

		tx.End()
	}
}

func (tw *genericWatcher) notify(target WatchTarget, ctx context.Context, wg *sync.WaitGroup, prevStatus ProductStatus, currStatus ProductStatus) {
	tw.dataChan <- &executionResult{
		target:     target,
		ctx:        ctx,
		wg:         wg,
		prevStatus: prevStatus,
		currStatus: currStatus,
	}
}

//
//func (tw *genericWatcher) GetProductUrl() string {
//	return tw.productUrl
//}
//
//func (tw *genericWatcher) GetWebhookPayload() WebhookPayload {
//	return tw.executionResult.product
//}

func (tw *genericWatcher) Stop() {
	tw.cancelFunc()
}
