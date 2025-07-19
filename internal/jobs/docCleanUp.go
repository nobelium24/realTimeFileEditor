package jobs

import (
	"context"
	"fmt"
	"log"
	"realTimeEditor/internal/model"
	"realTimeEditor/internal/repositories"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type DocumentCleanup struct {
	DocumentMedia repositories.DocumentMediaRepository
	cron          *cron.Cron
}

func NewReceiptCleanup(documentMedia repositories.DocumentMediaRepository) *DocumentCleanup {
	return &DocumentCleanup{
		DocumentMedia: documentMedia,
		cron:          cron.New(cron.WithSeconds()),
	}
}

func (r *DocumentCleanup) Start(ctx context.Context) {
	_, err := r.cron.AddFunc("0 */20 * * * *", func() {
		r.CleanupBatch(ctx)
	})
	if err != nil {
		log.Printf("Failed to schedule receipt cleanup: %v", err)
		return
	}

	r.cron.Start()
	go func() {
		<-ctx.Done()
		log.Println("Stopping receipt cleanup scheduler...")
		r.cron.Stop()
	}()
}

func (r *DocumentCleanup) CleanupBatch(ctx context.Context) {
	documentMedia, err := r.DocumentMedia.GetExpiredReceipts(20 * time.Minute)
	if err != nil {
		log.Printf("Error fetching receipts: %v", err)
		return
	}

	if len(documentMedia) == 0 {
		log.Printf("Nothing to cleanup, exiting...")
		return
	}
	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		deleteErr []string
	)

	semaphore := make(chan struct{}, 10)

	for _, documentMedium := range documentMedia {
		select {
		case <-ctx.Done():
			log.Println("Aborting cleanup batch due to shutdown")
			return
		default:
		}

		wg.Add(1)
		semaphore <- struct{}{}
		go func(medium model.DocumentMedia) {
			defer wg.Done()
			defer func() { <-semaphore }()

			if _, err := repositories.CloudinaryDelete(medium.PublicID, repositories.RawResource); err != nil {
				mu.Lock()
				deleteErr = append(deleteErr, fmt.Sprintf("Failed to delete %s: %v", medium.PublicID, err))
				mu.Unlock()
			}

			if err := r.DocumentMedia.Delete(medium.ID); err != nil {
				mu.Lock()
				deleteErr = append(deleteErr, fmt.Sprintf("Failed to delete %s: %v", medium.PublicID, err))
				mu.Unlock()
			}

		}(documentMedium)
	}

	wg.Wait()
	if len(deleteErr) > 0 {
		log.Printf("Cleanup completed with %d errors", len(deleteErr))
	}
}
