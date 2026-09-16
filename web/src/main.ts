import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import { applyTheme, currentTheme } from './stores/theme'
import { initHighlightTheme } from './utils/highlightTheme'

applyTheme(currentTheme.value)
initHighlightTheme()

createApp(App).mount('#app')
