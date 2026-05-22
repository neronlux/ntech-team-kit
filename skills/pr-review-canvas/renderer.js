function toggle(hdr) {
  var body = hdr.nextElementSibling, chev = hdr.querySelector('.chev');
  body.classList.toggle('open'); chev.classList.toggle('open');
}

function isImport(line) {
  var trimmed = line.replace(/^[+ -]/, '').trim();
  return trimmed.startsWith('import ') || trimmed.startsWith('import{') || trimmed.startsWith('} from ');
}

function isWhitespaceOnly(del, add) {
  return del.replace(/^-/, '').replace(/\s/g, '') === add.replace(/^\+/, '').replace(/\s/g, '');
}

function esc(s) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

function normWs(s) {
  return s.replace(/\s+/g, ' ').trim();
}

function toLines(input) {
  if (!input) return [];
  if (Array.isArray(input)) return input;
  if (typeof input === 'string') return input.split('\n');
  return [];
}

function loadPrDiffs() {
  var el = document.getElementById('pr-diffs-json');
  if (!el) return null;
  try {
    return JSON.parse(el.textContent || '');
  } catch (err) {
    console.error('Failed to parse PR diff JSON payload', err);
    return null;
  }
}

function detectMoves(deletions, additions) {
  var MIN_BLOCK = 3;
  var movedDels = {}, movedAdds = {};

  for (var delIdx = 0; delIdx < deletions.length; delIdx++) {
    if (movedDels[delIdx]) continue;

    var delBlock = [delIdx];
    for (var d = delIdx + 1; d < deletions.length && d - delIdx < 40; d++) {
      if (deletions[d].consecutive && !movedDels[d]) delBlock.push(d); else break;
    }
    if (delBlock.length < MIN_BLOCK) continue;

    var delNorms = delBlock.map(function(i) { return normWs(deletions[i].code); });

    for (var addIdx = 0; addIdx < additions.length; addIdx++) {
      if (movedAdds[addIdx]) continue;

      var addBlock = [addIdx];
      for (var a = addIdx + 1; a < additions.length && a - addIdx < 40; a++) {
        if (additions[a].consecutive && !movedAdds[a]) addBlock.push(a); else break;
      }
      if (addBlock.length < MIN_BLOCK) continue;

      var addNorms = addBlock.map(function(i) { return normWs(additions[i].code); });
      var matchLen = Math.min(delNorms.length, addNorms.length), matchCount = 0;
      for (var m = 0; m < matchLen; m++) { if (delNorms[m] === addNorms[m]) matchCount++; }

      if (matchCount >= MIN_BLOCK && matchCount >= matchLen * 0.7) {
        for (var k = 0; k < matchLen; k++) {
          movedDels[delBlock[k]] = { exact: delNorms[k] === addNorms[k] };
          movedAdds[addBlock[k]] = { exact: delNorms[k] === addNorms[k] };
        }
        break;
      }
    }
  }
  return { movedDels: movedDels, movedAdds: movedAdds };
}

function renderDiff(target, diffInput) {
  var el;
  if (typeof target === 'string') {
    el = document.getElementById(target) || document.querySelector(target);
  } else {
    el = target;
  }
  if (!el) return;

  var lines = toLines(diffInput);
  if (!lines.length) { el.innerHTML = '<div style="padding:12px;color:#777;font-size:12px;">No diff data</div>'; return; }

  var filtered = lines.filter(function(line) {
    if (line.startsWith('--- ') || line.startsWith('+++ ') || line.startsWith('@@') || line.startsWith('diff ')) return true;
    return !isImport(line);
  });

  var wsFiltered = [];
  for (var fi = 0; fi < filtered.length; fi++) {
    if (filtered[fi].startsWith('-')) {
      var delRun = [filtered[fi]], dj = fi + 1;
      while (dj < filtered.length && filtered[dj].startsWith('-')) { delRun.push(filtered[dj]); dj++; }
      var addRun = [], aj = dj;
      while (aj < filtered.length && filtered[aj].startsWith('+')) { addRun.push(filtered[aj]); aj++; }
      if (delRun.length === addRun.length && delRun.length > 0) {
        var allWs = true;
        for (var wc = 0; wc < delRun.length; wc++) { if (!isWhitespaceOnly(delRun[wc], addRun[wc])) { allWs = false; break; } }
        if (allWs) { for (var wx = 0; wx < addRun.length; wx++) wsFiltered.push(' ' + addRun[wx].slice(1)); fi = aj - 1; continue; }
      }
    }
    wsFiltered.push(filtered[fi]);
  }

  var deletions = [], additions = [], parsed = [];
  var oldLine = 0, newLine = 0, prevDel = false, prevAdd = false;
  for (var pi = 0; pi < wsFiltered.length; pi++) {
    var line = wsFiltered[pi];
    if (line.startsWith('--- ') || line.startsWith('+++ ') || line.startsWith('diff ')) continue;
    if (line.startsWith('@@')) {
      var hunkMatch = line.match(/@@ -(\d+)(?:,\d+)? \+(\d+)/);
      if (hunkMatch) { oldLine = parseInt(hunkMatch[1]); newLine = parseInt(hunkMatch[2]); }
      parsed.push({ type: 'hunk', text: line }); prevDel = false; prevAdd = false; continue;
    }
    if (line.startsWith('+')) {
      var addEntry = { type:'add', code:line.slice(1), newLine:newLine, consecutive:prevAdd, idx:parsed.length };
      additions.push(addEntry); parsed.push(addEntry); newLine++; prevAdd = true; prevDel = false;
    } else if (line.startsWith('-')) {
      var delEntry = { type:'del', code:line.slice(1), oldLine:oldLine, consecutive:prevDel, idx:parsed.length };
      deletions.push(delEntry); parsed.push(delEntry); oldLine++; prevDel = true; prevAdd = false;
    } else {
      var ctxCode = line.startsWith(' ') ? line.slice(1) : line;
      parsed.push({ type:'ctx', code:ctxCode, oldLine:oldLine, newLine:newLine }); oldLine++; newLine++; prevDel = false; prevAdd = false;
    }
  }

  var addIdxMap = new Map();
  for (var mi = 0; mi < additions.length; mi++) addIdxMap.set(additions[mi].idx, mi);
  var delIdxMap = new Map();
  for (var di = 0; di < deletions.length; di++) delIdxMap.set(deletions[di].idx, di);

  var moves = detectMoves(deletions, additions);
  var rows = [];
  for (var ri = 0; ri < parsed.length; ri++) {
    var entry = parsed[ri];
    if (entry.type === 'hunk') {
      rows.push('<tr class="diff-hunk"><td class="diff-ln"></td><td class="diff-ln"></td><td class="diff-code">' + esc(entry.text) + '</td></tr>');
    } else if (entry.type === 'add') {
      var addMapIdx = addIdxMap.get(entry.idx);
      var cls = (moves.movedAdds[addMapIdx]) ? (moves.movedAdds[addMapIdx].exact ? 'diff-moved-add' : 'diff-moved-add-edited') : 'diff-add';
      rows.push('<tr class="'+cls+'"><td class="diff-ln"></td><td class="diff-ln">'+entry.newLine+'</td><td class="diff-code">'+esc(entry.code)+'</td></tr>');
    } else if (entry.type === 'del') {
      var delMapIdx = delIdxMap.get(entry.idx);
      var cls2 = (moves.movedDels[delMapIdx]) ? (moves.movedDels[delMapIdx].exact ? 'diff-moved-del' : 'diff-moved-del-edited') : 'diff-del';
      rows.push('<tr class="'+cls2+'"><td class="diff-ln">'+entry.oldLine+'</td><td class="diff-ln"></td><td class="diff-code">'+esc(entry.code)+'</td></tr>');
    } else {
      rows.push('<tr class="diff-ctx"><td class="diff-ln">'+entry.oldLine+'</td><td class="diff-ln">'+entry.newLine+'</td><td class="diff-code">'+esc(entry.code)+'</td></tr>');
    }
  }
  el.innerHTML = '<table class="diff-table"><tbody>' + rows.join('') + '</tbody></table>';
}

document.addEventListener('DOMContentLoaded', function() {
  var prDiffs = loadPrDiffs();
  if (!prDiffs) return;
  var els = document.querySelectorAll('[data-diff]');
  for (var i = 0; i < els.length; i++) {
    var key = els[i].getAttribute('data-diff');
    if (key && Object.prototype.hasOwnProperty.call(prDiffs, key)) {
      renderDiff(els[i], prDiffs[key]);
    }
  }
});
