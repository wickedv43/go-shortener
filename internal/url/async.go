package url

import (
	"fmt"
	"sync"
	"time"
)

const batchSize = 10

// Gen takes a list of short URLs and returns a channel that emits them one by one.
func (u *URLService) Gen(shorts ...string) chan string {
	outCh := make(chan string)
	go func() {
		defer close(outCh)
		for _, short := range shorts {
			outCh <- short
		}
	}()

	return outCh
}

// Delete consumes short URLs from the input channel in batches and deletes them.
// Deletion happens either when the Batch size reaches a threshold or a timer fires.
// Returns a channel that emits status messages about the deletion process.
func (u *URLService) Delete(inCh chan string) chan string {
	outCh := make(chan string)

	timer := time.NewTicker(2 * time.Second)
	defer timer.Stop()

	go func() {
		defer close(outCh)
		var batch []string

		for {
			select {
			case short, ok := <-inCh:
				if !ok {
					//chan closed
					if len(batch) > 0 {
						err := u.BatchDelete(batch)
						if err != nil {
							u.logger.Errorf("failed to delete Batch: %v", err)
						}

						outCh <- fmt.Sprintf("deleted %d urls", len(batch))

					}
					return
				}

				batch = append(batch, short)

				if len(batch) >= batchSize {
					err := u.BatchDelete(batch)
					if err != nil {
						u.logger.Errorf("failed to delete Batch: %v", err)
					}

					outCh <- fmt.Sprintf("deleted %d urls", len(batch))

					batch = []string{}
				}

			case <-timer.C:
				if len(batch) > 0 {
					err := u.BatchDelete(batch)
					if err != nil {
						u.logger.Errorf("failed to delete Batch: %v", err)
					}

					outCh <- fmt.Sprintf("deleted %d urls", len(batch))

					batch = []string{}
				}
			}
		}
	}()

	return outCh
}

// FanIn merges multiple input channels into a single output channel.
// Useful for combining results from concurrent goroutines.
func (u *URLService) FanIn(chs ...chan string) chan string {
	var wg sync.WaitGroup
	outCh := make(chan string)

	output := func(c chan string) {
		for n := range c {
			outCh <- n
		}
		wg.Done()
	}

	wg.Add(len(chs))

	for _, c := range chs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(outCh)
	}()

	return outCh
}
