import { createRouter, createWebHistory } from 'vue-router';
import { HomeIcon } from '@hugeicons/core-free-icons';
import { PinIcon } from '@hugeicons/core-free-icons';
import { Settings01Icon } from '@hugeicons/core-free-icons';

import Welcome from '../resources/vue/views/Welcome.vue';
import ListView from '../resources/vue/views/ListView.vue';
import Options from '../resources/vue/views/Options.vue';

const routes = [
  {
    name: 'Welcome',
    path: '/',
    component: Welcome,
    meta: {
      title: 'Welcome',
      icon: HomeIcon,
    },
  },
  {
<<<<<<< HEAD
    name: 'ListView',
=======
    name: 'List',
>>>>>>> 71b59258622895210856dfda62060a71fc0bc8dc
    path: '/lists/:listId',
    component: ListView,
    props: true,
    meta: {
      title: 'List',
      icon: PinIcon,
      breadcrumbs: (route) => {
        return [
          { name: 'Home', href: '/' },
          { name: 'List: ' + String(route.params.listId), href: '/lists/' + route.params.listId }
        ];
      }
    }
  },
  {
    name: 'Search',
    path: '/search',
    component: ListView,
    props: (route) => ({ listId: null, searchQuery: route.query.q || '' }),
    meta: {
      title: 'Search',
      icon: PinIcon,
      breadcrumbs: (route) => {
        return [
          { name: 'Home', to: '/' },
          { name: 'Search: ' + (route.query.q || ''), to: '/search' }
        ];
      }
    }
  },
  {
    name: 'Options',
    path: '/options',
    component: Options,
    meta: {
      title: 'Options',
      icon: Settings01Icon,
    }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
