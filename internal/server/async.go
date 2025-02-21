package server

import (
	"fmt"
	"sync"
	"time"
)

const batchSize = 10

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
							s.logger.Errorf("failed to delete batch: %v", err)
						}
						for _, url := range batch {
							outCh <- fmt.Sprintf("deleted: %s", url)
						}
					}
					return
				}

				batch = append(batch, short)

				// Если набралось `batchSize`, отправляем в `batchDelete`
				if len(batch) >= batchSize {
					err := s.batchDelete(batch)
					if err != nil {
						s.logger.Errorf("failed to delete batch: %v", err)
					}

					for _, url := range batch {
						outCh <- fmt.Sprintf("deleted: %s", url)
					}
					batch = nil // Очищаем batch
				}

			case <-timer.C:
				if len(batch) > 0 {
					err := s.batchDelete(batch)
					if err != nil {
						s.logger.Errorf("failed to delete batch: %v", err)
					}

					for _, url := range batch {
						outCh <- fmt.Sprintf("deleted: %s", url)
					}
					batch = nil
				}
			}
		}
	}()

	return outCh
}

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
