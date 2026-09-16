import { watch } from 'vue';
import darkCss from 'highlight.js/styles/github-dark.css?inline';
import lightCss from 'highlight.js/styles/github.css?inline';
import { currentTheme } from '../stores/theme';

const STYLE_ID = 'hljs-theme';

export function initHighlightTheme(): void {
  const style = document.createElement('style');
  style.id = STYLE_ID;
  document.head.appendChild(style);
  watch(
    currentTheme,
    (theme) => {
      style.textContent = theme === 'light' ? lightCss : darkCss;
    },
    { immediate: true },
  );
}
