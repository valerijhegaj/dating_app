package indexer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"date-app/configs"
	"date-app/internal/storage"
)

func New(data storage.Storage) Indexer {
	return Indexer{startShutdownChan: make(chan struct{}), data: data}
}

type Indexer struct {
	startShutdownChan chan struct{}
	data              storage.Storage
}

var t time.Time
var loc *time.Location

func init() {
	var err error
	if t, err = time.Parse(
		"15:04", configs.Config.Main.TimeToIndex,
	); err != nil {
		log.Fatal(err)
	}
	if loc, err = time.LoadLocation("Local"); err != nil {
		log.Fatal(err)
	}

}

func (i *Indexer) timeToIndex() <-chan time.Time {
	now := time.Now().Local()
	callTime := time.Date(
		now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0,
		0, loc,
	)
	if callTime.Before(now) {
		callTime = callTime.Add(time.Hour * 24)
	}
	return time.After(callTime.Sub(now))
}

func (i *Indexer) Start(finalizeShutdownChan chan struct{}) {
	go func() {
	Loop:
		for {
			select {
			case <-i.startShutdownChan:
				break Loop
			case <-i.timeToIndex():
			}
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				select {
				case <-i.startShutdownChan:
					cancel()
				case <-ctx.Done():
				}
			}()
			if err := i.indexing(ctx); err != nil {
				log.Printf("Indexing error: %w", err)
			}
			cancel()
		}
		log.Println("Indexer stopped.")
		finalizeShutdownChan <- struct{}{}
	}()
}

func (i *Indexer) IndexUser(ctx context.Context, userID int) error {
	likers, err := i.data.GetNewLikes(ctx, userID)
	if err != nil {
		return err
	}
	indexedUserID, err := i.data.GetNewPairs(
		ctx, userID, configs.Config.Main.NumToIndex,
	)
	if err != nil {
		return err
	}
	indexedUserID = append(likers, indexedUserID...)
	if err = i.data.LoadIndexed(
		ctx, userID, indexedUserID,
	); err != nil {
		return err
	}
	return nil
}

func (i *Indexer) indexing(ctx context.Context) error {
	const op = "indexer.indexing"

	userIDs, err := i.data.GetAllUserIDs(ctx)
	if err != nil {
		return fmt.Errorf("%v: %w", op, err)
	}
	var errs []error
	for _, userID := range userIDs {
		if err = i.IndexUser(ctx, userID); err != nil {
			log.Printf("%v: %w", op, err)
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return fmt.Errorf(
			"%v: %w", op, errors.New("indexing errors check logs"),
		)
	}
	return nil
}

func (i *Indexer) Shutdown(shutdownCtx context.Context) error {
	select {
	case <-shutdownCtx.Done():
		return errors.New("indexer didn't shutdown ctx canceled")
	case i.startShutdownChan <- struct{}{}:
		return nil
	}
}
