import { ScpiNode } from "./types";

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

export function childrenOf(input: ScpiNode[]): ScpiNode[] {
  const all: ScpiNode[] = [];
  for (const node of input) {
    if (node.children) {
      all.push(...node.children);
    }
  }
  return all;
}
export function removeDuplicateNodes(input: ScpiNode[]): ScpiNode[] {
  const seen = new Map<string, boolean>();
  return input
    .filter(item => {
      const key = JSON.stringify(item.content);
      if (seen.has(key)) return false;
      seen.set(key, true);
      return true;
    }).sort((a, b) => {
      const textCompare = a.content.text.localeCompare(b.content.text);
      if (textCompare !== 0) return textCompare;
      return (a.content.suffixed ? 1 : 0) - (b.content.suffixed ? 1 : 0);
    });
}

export function findCardinalNode(input: ScpiNode[]): ScpiNode | undefined {
  for (const node of input) {
    if (node.content.suffixed) {
      return node;
    }
  }
  return undefined;
}

export async function getClipboardText(): Promise<string> {
  try {
    const text = await navigator.clipboard.readText();
    return text;
  } catch (err) {
    return ''
  }
}

export function getTimestamp(): string {
  const now = new Date();
  return now.toISOString().replace(/[:.]/g, '-').slice(0, -5);
}

