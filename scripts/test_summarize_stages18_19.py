"""Regression for null idle samples and timestamped nearest-rank summaries."""
import json,pathlib,subprocess,sys,tempfile,unittest

class SummaryTest(unittest.TestCase):
 def test_idle_and_tail_statistics(self):
  with tempfile.TemporaryDirectory(prefix='nesting-summary-') as work:
   root=pathlib.Path(work);rows=[{'Endpoint':'POST /projects/1/surfaces/1/nesting','Status':status,'Error':False,'Milliseconds':value,'StartedUnixMS':100000+i*1000,'Algorithm':'baseline'} for i,(status,value) in enumerate([(200,1),(200,2),(503,100)])]
   phases=[{'Name':'soak','Clients':1,'APIs':1,'StartedUnixMS':100000,'Seconds':60,'Samples':rows,'Diagnostics':[]},{'Name':'idle','Clients':0,'APIs':1,'StartedUnixMS':160000,'Seconds':60,'Samples':None,'Diagnostics':[]}]
   log='\n'.join('LOAD_PHASE_JSON '+json.dumps(p) for p in phases)+'\n--- PASS: TestStage18MixedLoad/soak (60s)\n--- PASS: TestStage18MixedLoad/idle (60s)\n'
   (root/'test.log').write_text(log)
   subprocess.run([sys.executable,str(pathlib.Path(__file__).with_name('summarize_stages18_19.py')),str(root)],check=True,stdout=subprocess.DEVNULL)
   out=json.loads((root/'summary.json').read_text());active,idle=out['phases']
   self.assertEqual((active['p50'],active['p95'],active['p99']),(2,100,100));self.assertEqual(active['requests'],3)
   self.assertEqual(idle['requests'],0);self.assertEqual(idle['rps'],0);self.assertIsNone(idle['p99']);self.assertTrue(idle['test_passed'])
   self.assertEqual(out['latency_windows'][0]['unexpected_5xx'],0);self.assertTrue(out['latency_windows'][0]['p99_small_sample']);self.assertEqual(out['http_by_algorithm'][0]['statuses'],{'200':2,'503':1})

if __name__=='__main__':unittest.main()
