export const range = (start: number, stop: number) =>
  Array.from({ length: stop - start + 1 }, (_, i) => start + i);

export function getShortMnemonic(input: string): string {
  return input.replace(/[a-z]/g, '')
}

export function stripCardinality(input: string): string {
  const match = input.match(/^[a-zA-Z]+/);
  const isQuery = input.endsWith('?');
  if (match && isQuery) {
    return match[0] + '?';
  } else if (match && !isQuery) {
    return match[0];
  } else {
    return '';
  }
}

export function cardinalityOf(input: string): number | undefined {
  const match = input.match(/\d+$/);
  return match ? Number(match[0]) : undefined;
}
