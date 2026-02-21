import { createRouter, createWebHistory } from 'vue-router';
import { HomeIcon } from '@hugeicons/core-free-icons';
import { PinIcon } from '@hugeicons/core-free-icons';
import { Settings01Icon } from '@hugeicons/core-free-icons';

import Welcome from '../resources/vue/views/Welcome.vue';
import ListView from '../resources/vue/views/ListView.vue';
import NavOptions from '../resources/vue/components/NavOptions.vue';
import Options from '../resources/vue/views/Options.vue';
import Diagnostics from '../resources/vue/views/Diagnostics.vue';

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
    name: 'ListView',
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
    name: 'NavOptions',
    path: '/options',
    component: NavOptions,
    meta: {
      title: 'Options',
      icon: Settings01Icon,
    }
  },
  {
    name: 'Settings',
    path: '/settings',
    component: Options,
    meta: {
      title: 'Settings',
      icon: Settings01Icon,
    }
  },
  {
    name: 'Diagnostics',
    path: '/diagnostics',
    component: Diagnostics,
    meta: {
      title: 'Diagnostics',
      icon: Settings01Icon,
    }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
