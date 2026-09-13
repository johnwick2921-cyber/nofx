#!/usr/bin/env python3
"""Build an offline behavioral harness from the actual production method bodies.
NT8 is replaced by inert recording fakes; no live assemblies/connections are used.
Pass an output .cs path. Compile with framework csc or Roslyn + framework refs.
"""
from pathlib import Path
import re,sys
src=(Path(__file__).parents[1]/'VLTraderTCPClient.cs').read_text()
names=['TryResolveExecutionAccount','TryResolveHeldPosition','HandleClosePosition','HandlePlaceProtectiveStop','HandleCancelOrder','OnOrderUpdate','SubmitBracketOnEntryFill','AmendBracketQuantity','IsLiveAtExchange','IsTerminalOrderState','RetireTerminalBracket','CancelBracketsFor','CancelAllBracketsFor','GetString','GetDouble','GetInt']
def member(name):
 m=re.search(r'^        private (?:static )?[^\n]+\b'+name+r'\(',src,re.M)
 if not m:raise RuntimeError(name)
 end=re.search(r'^        }',src[m.end():],re.M)
 return src[m.start():m.end()+end.end()]+'\n'
classes=[]
for name in ['PendingBracket','PlacedBracket']:
 m=re.search(r'^        private class '+name+r'\b',src,re.M);end=re.search(r'^        }',src[m.end():],re.M);classes.append(src[m.start():m.end()+end.end()])
template=(Path(__file__).parent/'lifecycle_harness.cs.in').read_text()
Path(sys.argv[1]).write_text(template.replace('/* PRODUCTION TYPES */','\n'.join(classes)).replace('/* PRODUCTION METHODS */','\n'.join(member(n) for n in names)))
print('Extracted '+str(len(names))+' production methods into '+sys.argv[1])
