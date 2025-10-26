import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
  name: 'syntaxHighlight'
})
export class SyntaxHighlightPipe implements PipeTransform {

  transform(value: string | undefined, ...args: unknown[]): string | undefined {
    if (!value) return value;

    let result = '';
    let i = 0;

    while (i < value.length) {
      const char = value[i];

      if (char === ':') {
        result += '<span class="colon">:</span>';
        i++;
      } else if (char === '*') {
        result += '<span class="star">*</span>';
        i++;
      } else if (char === '?') {
        result += '<span class="question-mark">?</span>';
        i++;
      } else if (/[A-Z]/.test(char)) {
        let upperSequence = '';
        while (i < value.length && /[A-Z]/.test(value[i])) {
          upperSequence += value[i];
          i++;
        }
        result += `<span class="short">${upperSequence}</span>`;
      } else {
        result += char;
        i++;
      }
    }

    return result;
  }

}
