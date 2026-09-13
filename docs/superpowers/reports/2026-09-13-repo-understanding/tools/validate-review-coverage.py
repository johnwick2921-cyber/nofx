"""Validate read-ledger provenance and named Go declaration census.

This verifies artifact consistency, not whether a human/model understood code.
Run from the pinned worktree, passing the audit report directory.
"""
import csv
import json
import pathlib
import sys

root = pathlib.Path(sys.argv[1])
plan = json.loads((root / 'review-plan.json').read_text())
declarations = list(csv.DictReader((root / 'go-functions.tsv').open(), delimiter='\t'))
results = []
for assignment in plan['assignments']:
    report = root / 'reviews' / f"{assignment['id']:02}"
    if not (report / 'reads.json').exists():
        continue
    reads = json.loads((report / 'reads.json').read_text())
    by_path = {f['path']: f for f in reads['files']}
    failures = []
    if reads['base'] != plan['base']:
        failures.append('base mismatch')
    for source in assignment.get('files', []):
        read = by_path.get(source['path'])
        if not read or read.get('mode') != 'full/manual':
            failures.append('not full: ' + source['path'])
            continue
        if read['sha256'] != source['sha256']:
            failures.append('hash mismatch: ' + source['path'])
        covered = set()
        for start, end in read['ranges']:
            covered.update(range(start, end + 1))
        if not set(range(1, source['lines'] + 1)).issubset(covered):
            failures.append('line gaps: ' + source['path'])
    notes = json.loads((report / 'functions.json').read_text())
    if isinstance(notes, dict):
        notes = notes['functions']
    # Some workers preserve receiver-qualified names; the AST census keeps
    # receiver in a separate column. Source path + declaration line disambiguate.
    noted = {(n['path'], n['name'].rsplit('.', 1)[-1], int(n['start_line'])) for n in notes}
    paths = {f['path'] for f in assignment.get('files', [])}
    expected = {(f['path'], f['name'], int(f['start_line'])) for f in declarations
                if f['kind'] == 'declaration' and f['path'] in paths}
    missing = sorted(expected - noted)
    results.append({'assignment': assignment['id'], 'files': len(paths),
                    'ledger_errors': failures, 'named_go_declarations': len(expected),
                    'missing_named_go_notes': missing,
                    'warning': 'Consistency only; not independent semantic or runtime verification.'})
(root / 'coverage-validation.json').write_text(json.dumps(results, indent=2) + '\n')
print(json.dumps([{'assignment':r['assignment'], 'ledger_errors':len(r['ledger_errors']),
                   'named_go_declarations':r['named_go_declarations'],
                   'missing_named_go_notes':len(r['missing_named_go_notes'])} for r in results], indent=2))
sys.exit(any(r['ledger_errors'] or r['missing_named_go_notes'] for r in results))
