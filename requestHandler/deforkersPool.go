package requestHandler

import "sync"

type CompleteFile FileLink

type DeforkersPool struct {
	In1              chan *DownloadedFile
	In2              chan *RecognizedPhoto
	Out              chan *CompleteFile
	DeforkDownloaded FromDownloadedToComplete
	DeforkRecognized FromRecognizedToComplete
	Done             chan struct{}
	WorkersNumber    int
}

type FromDownloadedToComplete func(*DownloadedFile) (*CompleteFile, error)

type FromRecognizedToComplete func(*RecognizedPhoto) (*CompleteFile, error)

func (dp *DeforkersPool) Run(queueSize int, finished sync.WaitGroup) chan *CompleteFile {
	dp.Out = make(chan *CompleteFile, queueSize)
	var wg sync.WaitGroup
	wg.Add(dp.WorkersNumber)
	for i := 0; i < dp.WorkersNumber; i++ {
		go func() {
			dp.defork()
			wg.Done()
		}()
	}
	finished.Add(1)
	go func() {
		wg.Wait()
		close(dp.Out)
		finished.Done()
	}()
	return dp.Out
}

func (dp *DeforkersPool) defork() {
	for {
		select {
		case in1 := <-dp.In1:
			appContext.SysLogger.Printf("comes on defork1: %+v", *in1)
			out, err := dp.DeforkDownloaded(in1)
			if err != nil {
				continue
			}

			appContext.SysLogger.Printf("comes from defork1: %+v", *out)
			select {
			case dp.Out <- out:
			case <-dp.Done:
				return
			}
		case in2 := <-dp.In2:
			appContext.SysLogger.Printf("comes on defork2: %+v", *in2)
			out, err := dp.DeforkRecognized(in2)
			if err != nil {
				continue
			}

			appContext.SysLogger.Printf("comes from defork2: %+v", *out)
			select {
			case dp.Out <- out:
			case <-dp.Done:
				return
			}
		}
	}
}
