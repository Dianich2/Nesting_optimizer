"""Build a test binary with a Go overlay for bounded limiter diagnostics.
The production file is READ only. Original channel/transaction/limit logic stays
unchanged; overlay adds atomic counters and 32 fixed histogram buckets. Keep the
JSON and overlaid source alongside results to disclose observer overhead.
"""
import argparse,json,pathlib,subprocess
p=argparse.ArgumentParser();p.add_argument('--backend',type=pathlib.Path,required=True);p.add_argument('--output',type=pathlib.Path,required=True);p.add_argument('--evidence',type=pathlib.Path,required=True);a=p.parse_args();a.evidence.mkdir(parents=True,exist_ok=True)
original=(a.backend/'internal/usecase/nesting/limiter.go').resolve();s=original.read_text()
s=s.replace('"context"','"context"\n"sync/atomic"\n"time"\n"math/bits"',1)
needle='type ComputeLimiter struct {';assert s.count(needle)==1
s=s.replace(needle,needle+'''
 labQueue [32]atomic.Int64
 labCalls atomic.Int64
 labRejects atomic.Int64
 labQueueNs atomic.Int64
 labQueueMax atomic.Int64
 labAdmissionNs atomic.Int64
''',1)
start=s.index('func (l *ComputeLimiter) Acquire');head,body=s[:start],s[start:]
needle='case l.admitted <- struct{}{}:';assert body.count(needle)==1
body=body.replace(needle,'case l.admitted <- struct{}{}:\n l.labAdmissionNs.Add(time.Since(labStart).Nanoseconds())',1)
body=body.replace('if err := ctx.Err();','labStart:=time.Now()\n if err := ctx.Err();',1)
body=body.replace('default:\n\t\treturn nil,','default:\n l.labRejects.Add(1)\n\t\treturn nil,',1)
needle='\tselect {\n\tcase l.running';assert needle in body
body=body.replace(needle,'\tlabQueued:=time.Now()\n\tselect {\n\tcase l.running',1)
body=body.replace('case l.running <- struct{}{}:','case l.running <- struct{}{}:\n l.labRecordQueue(time.Since(labQueued))',1)
body=body.replace('case <-ctx.Done():\n\t\t<-l.admitted','case <-ctx.Done():\n l.labRecordQueue(time.Since(labQueued))\n\t\t<-l.admitted',1)
s=head+body+'''
// TEST OVERLAY ONLY: no method or diagnostic counters are shipped in cmd/api.
func(l *ComputeLimiter) labRecordQueue(d time.Duration) {
 n:=d.Nanoseconds();l.labCalls.Add(1);l.labQueueNs.Add(n)
 bucket:=bits.Len64(uint64(n/1000+1));if bucket>31{bucket=31};l.labQueue[bucket].Add(1)
 for {old:=l.labQueueMax.Load();if n<=old||l.labQueueMax.CompareAndSwap(old,n){break}}
}
func(l *ComputeLimiter) LabMetrics() map[string]any {
 buckets:=make([]int64,32);for i:=range buckets {buckets[i]=l.labQueue[i].Load()}
 return map[string]any{"acquired_or_cancelled":l.labCalls.Load(),"rejected":l.labRejects.Load(),"queue_sum_ns":l.labQueueNs.Load(),"queue_max_ns":l.labQueueMax.Load(),"admission_sum_ns":l.labAdmissionNs.Load(),"queue_log2_us_buckets":buckets}
}
'''
source=(a.evidence/'limiter-observed.go.txt').resolve();source.write_text(s)
subprocess.run(['gofmt','-w',str(source)],check=True)
run_original=(a.backend/'internal/usecase/nesting/run.go').resolve()
r=run_original.read_text().replace('"context"','"context"\n"sync/atomic"\n"runtime/pprof"',1)
needle='type RunNestingUseCase struct {';assert r.count(needle)==1
r=r.replace(needle,needle+'\nlabOptCalls [5]atomic.Int64\nlabOptNS [5]atomic.Int64\nlabOptMax [5]atomic.Int64',1)
needle='result, err := optimizer.Optimize(\n\t\tctx,\n\t\tproblem,\n\t)';assert r.count(needle)==1
r=r.replace(needle,'''var result nesting.Result
pprof.Do(ctx,pprof.Labels("algorithm",string(input.Algorithm)),func(labeled context.Context){result,err=optimizer.Optimize(labeled,problem)})
uc.labRecordOptimize(input.Algorithm,time.Since(startAt))''',1)
r+='''
// TEST OVERLAY ONLY: bounded five-algorithm diagnostics; production file untouched.
func (uc *RunNestingUseCase) labRecordOptimize(a nesting.Algorithm,d time.Duration) {
 i:=0;switch a {case nesting.NFPGreedyAlgorithm:i=1;case nesting.GeneticAlgorithm:i=2;case nesting.SimulatedAnnealingAlgorithm:i=3;case nesting.HybridAlgorithm:i=4}
 n:=d.Nanoseconds();uc.labOptCalls[i].Add(1);uc.labOptNS[i].Add(n)
 for{old:=uc.labOptMax[i].Load();if n<=old||uc.labOptMax[i].CompareAndSwap(old,n){break}}
}
func(uc *RunNestingUseCase) LabOptimizeMetrics()map[string]any {
 out:=map[string]any{};for i,name:=range []string{"baseline","nfp_greedy","genetic","simulated_annealing","hybrid"}{out[name]=map[string]int64{"calls":uc.labOptCalls[i].Load(),"elapsed_total_ns":uc.labOptNS[i].Load(),"elapsed_max_ns":uc.labOptMax[i].Load()}};return out
}
'''
run_source=(a.evidence/'run-observed.go.txt').resolve();run_source.write_text(r)
subprocess.run(['gofmt','-w',str(run_source)],check=True)
overlay=(a.evidence/'overlay.json').resolve();overlay.write_text(json.dumps({'Replace':{str(original):str(source),str(run_original):str(run_source)}}))
subprocess.run(['go','test','-p=2','-overlay',str(overlay),'-c','-o',str(a.output.resolve()),'./internal/integration'],cwd=a.backend,check=True)
print('Built test-only limiter/Optimize observer overlay; production sources unchanged.')
