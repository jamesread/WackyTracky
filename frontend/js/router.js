import { createRouter, createWebHistory } from 'vue-router';

import Welcome from '../resources/vue/views/Welcome.vue';
import ListView from '../resources/vue/views/ListView.vue';

const routes = [
  {
    path: '/',
    component: Welcome,
  },
  {
    path: '/lists/:listId',
    component: ListView,
    props: true,
    meta: {
      breadcrumbs: (route) => {
        return [
          { name: 'Home', to: '/' },
          { name: 'List: ' + String(route.params.listId), to: '/lists/:listId' }
        ];
      }
    }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
