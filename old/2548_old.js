  return type === 'band' || type === 'ordinal'
      || type === 'point' || type === 'bin-ordinal';
}

function domainCaption(scale) {
  const MAX = 10;
  const d = scale.domain();

  if (isDiscrete(scale.type)) {
    const n = d.length,
          v = n > MAX ? d.slice(0, MAX - 2).concat('\u2026', d.slice(-1)) : d;
    return ` with ${n} discrete value${n !== 1 ? 's' : ''}: ${v.join(', ')}`;
  } else {
    return ` with values from ${d[0]} to ${peek(d)}`;
  }
}
