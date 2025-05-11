package server

import (
	"fmt"
	"sync"
	"time"
)

const batchSize = 10

// gen takes a list of short URLs and returns a channel that emits them one by one.
func (s *Server) gen(shorts ...string) chan string {
	outCh := make(chan string)
	go func() {
		defer close(outCh)
		for _, short := range shorts {
			outCh <- short
		}
	}()

	return outCh
}

// delete consumes short URLs from the input channel in batches and deletes them.
// Deletion happens either when the Batch size reaches a threshold or a timer fires.
// Returns a channel that emits status messages about the deletion process.
func (s *Server) delete(inCh chan string) chan string {
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
						err := s.batchDelete(batch)
						if err != nil {
							s.logger.Errorf("failed to delete Batch: %v", err)
						}

						outCh <- fmt.Sprintf("deleted %d urls", len(batch))

					}
					return
				}

				batch = append(batch, short)

				if len(batch) >= batchSize {
					err := s.batchDelete(batch)
					if err != nil {
						s.logger.Errorf("failed to delete Batch: %v", err)
					}

					outCh <- fmt.Sprintf("deleted %d urls", len(batch))

					batch = []string{}
				}

			case <-timer.C:
				if len(batch) > 0 {
					err := s.batchDelete(batch)
					if err != nil {
						s.logger.Errorf("failed to delete Batch: %v", err)
					}

					outCh <- fmt.Sprintf("deleted %d urls", len(batch))

					batch = []string{}
				}
			}
		}
	}()

	return outCh
}

// fanIn merges multiple input channels into a single output channel.
// Useful for combining results from concurrent goroutines.
func (s *Server) fanIn(chs ...chan string) chan string {
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
