package retrieval_api

import (
	"container/heap"
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"github.com/miekg/dns"
	"io"
	"os"
	_ "runtime" // HEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHE
	"strconv"
	"strings"
	"sync"
	"time"
	_ "unsafe" // HEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHE
)

//go:embed lists/resolvers.txt
var resolversSource string

var (
	Resolvers         []string
	NumberOfResolvers int
	StatusOutput      io.Writer = os.Stderr
)

type cdn53RR struct {
	Domain         string
	Answer         string
	Response       bool
	Id             uint
	TargetResolver string
	Latency        time.Duration
	err            error
}

type cdn53Answers []*cdn53RR

func (c53a cdn53Answers) Len() int {
	return len(c53a)
}

func (c53a cdn53Answers) Swap(i, j int) {
	c53a[i], c53a[j] = c53a[j], c53a[i]
}

func (c53a cdn53Answers) Less(i, j int) bool {
	return c53a[i].Id < c53a[j].Id
}

func (h *cdn53Answers) Push(x any) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(*cdn53RR))
}

func (h *cdn53Answers) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func init() {
	Resolvers = strings.Split(resolversSource, "\n")
	NumberOfResolvers = len(Resolvers)
	// Fun fact, sometimes len returns -1. IDK why but this is a *huge* problem. So we're gonna
	// grab len here and use it later
}

func splitDomain(domain string) (string, string) {
	split := strings.SplitN(domain, ".", 2)
	return split[0], split[1]
}

func DownloadAndDecode(domain string, asyncLimit uint, perRequestTimeout time.Duration, output io.Writer) error {
	NumberOfResolvers = len(Resolvers)
	wg := sync.WaitGroup{}

	// DNS Answers get sent here
	rrAnswers := make(chan *cdn53RR, 200)

	// DNS Queries get sent here to be turned into answers by workers
	rrQueries := make(chan *cdn53RR, 200)

	// Answers get ordered and sent here so that they can be written and decoded to disk in order
	orderedAnswers := make(chan *cdn53RR, 200)

	dnsClient := new(dns.Client)
	// Ironically, most if not all resolvers refuse to return TXT records
	// that exceed 255 characters over UDP. We must use TCP or TCP-TLS.
	// When you think about it this makes sense because of MTU, etc. sizes
	dnsClient.Net = "tcp"

	// Start all our workers
	wg.Add(int(asyncLimit) + 1)
	go decodeAnswersToOutputWorker(orderedAnswers, output, wg.Done)

	for i := uint(0); i < asyncLimit; i++ {
		go queryDNSWithTimeoutWorker(rrQueries, dnsClient, &perRequestTimeout, rrAnswers, wg.Done)
	}

	leaf, root := splitDomain(domain)
	lastSegment := 0
	{
		query := &cdn53RR{}
		query.Domain = fmt.Sprintf("%s-0.%s", leaf, root)
		query.Id = 0

		rrQueries <- query
		answer := <-rrAnswers
		_, _ = fmt.Fprintf(StatusOutput, "Total segments: %v\n", answer.Answer)

		lastSegment, _ = strconv.Atoi(answer.Answer)
	}

	start := time.Now()
	queriedRequestIndex := uint(1)
	nextOrderedAnswerIndex := uint(1)
	taskCount := uint(0)
	lastSegmentId := uint(lastSegment)

	recordedAnswers := &cdn53Answers{}
	heap.Init(recordedAnswers)
	for queriedRequestIndex <= lastSegmentId {
		if taskCount == 0 {
			for taskCount < asyncLimit && queriedRequestIndex <= lastSegmentId {
				segmentQuery := &cdn53RR{
					Domain: fmt.Sprintf("%s-%v.%s", leaf, queriedRequestIndex, root),
					Id:     queriedRequestIndex,
				}
				rrQueries <- segmentQuery
				queriedRequestIndex++
				taskCount++
			}
		}

		// Wait for all of the answers in this taskGroup. This is done so that we can order and write all of
		// the elements in this task group before generating more of them. If we are pulling big data this
		// matters.
		for taskCount > 0 {
			select {
			case segmentAnswer := <-rrAnswers:
				taskCount--
				if segmentAnswer.Id == nextOrderedAnswerIndex {
					orderedAnswers <- segmentAnswer
					nextOrderedAnswerIndex++
					for len(*recordedAnswers) > 0 && (*recordedAnswers)[0] != nil && (*recordedAnswers)[0].Id == nextOrderedAnswerIndex {
						segmentAnswer = heap.Pop(recordedAnswers).(*cdn53RR)
						orderedAnswers <- segmentAnswer
						nextOrderedAnswerIndex++
					}

					_, _ = fmt.Fprintf(StatusOutput, "\r\033[2KDecoding and writing segment %v", segmentAnswer.Id)
				} else { // Store and order, then forward
					heap.Push(recordedAnswers, segmentAnswer)
				}

				if taskCount <= 0 {
					break
				}
			}
		}
	}

	close(rrQueries)
	close(rrAnswers)
	close(orderedAnswers)
	wg.Wait()
	_, _ = fmt.Fprintf(StatusOutput, "\nAll %v segments downloaded and decoded in %v.\n", lastSegment, time.Now().Sub(start))
	return nil
}

func decodeAnswersToOutputWorker(rrAnswers chan *cdn53RR, output io.Writer, done func()) {
	defer done()

	for {
		select {
		case answer, ok := <-rrAnswers:
			if ok {
				decodedAnswer, err := base64.StdEncoding.DecodeString(answer.Answer)
				if err != nil {
					panic(err)
				}

				_, _ = output.Write(decodedAnswer)
			} else { // Channel closed... Exit
				return
			}
		}
	}
}

func queryDNSWithTimeoutWorker(rrQueries chan *cdn53RR, client *dns.Client, timeout *time.Duration, rrAnswers chan *cdn53RR, done func()) {
	defer done()
	for {
		select {
		case query, ok := <-rrQueries:
			if ok {
				query.Response = false
				rrAnswer := query
				for rrAnswer.Response == false {
					query.TargetResolver = randomResolver()
					rrAnswer = queryDNSWithTimeout(query, client, timeout)
				}

				rrAnswers <- rrAnswer
			} else { // Channel closed... Exit
				return
			}
		}
	}
}

func queryDNSWithTimeout(query *cdn53RR, client *dns.Client, timeout *time.Duration) *cdn53RR {
	ctxt, cncl := context.WithTimeout(context.Background(), *timeout)
	defer cncl()

	txtReq := new(dns.Msg)
	txtReq.SetQuestion(dns.Fqdn(query.Domain), dns.TypeTXT)

	rsp, latency, err := client.ExchangeContext(ctxt, txtReq, query.TargetResolver)
	if err != nil {
		query.err = err
		return query
	}

	answerBuilder := strings.Builder{}
	for _, answer := range rsp.Answer {
		for _, rr := range answer.(*dns.TXT).Txt {
			rr = strings.ReplaceAll(rr, "\"", "")
			rr = strings.ReplaceAll(rr, " ", "")
			answerBuilder.WriteString(rr)
		}
	}

	answer := answerBuilder.String()

	query.err = err
	query.Answer = answer
	query.Response = len(answer) > 0 // We will always have data. The lack of data is not a response
	query.Latency = latency
	return query
}

// HEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHEHE
// So.. This is the most performant async safe PRNG in golang. I know it includes unsafe to use it... But I mean c'mon
// I'm not going to use a mutex just because rand isn't async-safe
//go:linkname fastrand runtime.fastrand
func fastrand() uint32

func randomResolver() string {
	num := fastrand() % uint32(NumberOfResolvers)

	return Resolvers[num] + ":53"
}
