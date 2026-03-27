import { createApp } from 'vue'
import App from './App.vue'
import router from './router';
import { createPinia } from 'pinia';
import 'katex/dist/katex.min.css';
import './styles/globals.css';

const app = createApp(App)
  .use(createPinia())
  .use(router);

router.isReady().then(() => {
  app.mount('#app');
});
