import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import { i18n } from './i18n'
import { router } from './router'
import 'virtual:uno.css'
import './styles/tokens.css'
import './styles/base.css'
import './styles/element-overrides.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(i18n)

app.mount('#app')
