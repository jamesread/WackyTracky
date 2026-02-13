import { createRouter, createWebHistory } from 'vue-router';

import Welcome from '../resources/vue/views/Welcome.vue';
import ListView from '../resources/vue/views/ListView.vue';

const routes = [
  {
    name: 'Welcome',
    path: '/',
    component: Welcome,
  },
  {
    name: 'List',
    path: '/lists/:listId',
    component: ListView,
    props: true,
    meta: {
      breadcrumbs: (route) => {
        return [
          { name: 'Home', href: '/' },
          { name: 'List: ' + String(route.params.listId), href: '/lists/' + route.params.listId }
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
