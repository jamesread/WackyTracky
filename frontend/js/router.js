import { createRouter, createWebHistory } from 'vue-router';
import { HomeIcon } from '@hugeicons/core-free-icons';
import { PinIcon } from '@hugeicons/core-free-icons';
import { Settings01Icon } from '@hugeicons/core-free-icons';

import Welcome from '../resources/vue/views/Welcome.vue';
import ListView from '../resources/vue/views/ListView.vue';
import NavOptions from '../resources/vue/components/NavOptions.vue';
import Options from '../resources/vue/views/Options.vue';
import Diagnostics from '../resources/vue/views/Diagnostics.vue';
import TaskPropertyProperties from '../resources/vue/views/TaskPropertyProperties.vue';

const routes = [
  {
    name: 'Welcome',
    path: '/',
    component: Welcome,
    meta: {
      title: 'Welcome',
      icon: HomeIcon,
      description: 'Choose a list to view or add tasks.',
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
      description: 'Search tasks across all lists by text, tags, or contexts.',
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
      description: 'Shortcuts to settings, task properties, and diagnostics.',
    }
  },
  {
    name: 'TaskPropertyProperties',
    path: '/options/task-property-properties',
    component: TaskPropertyProperties,
    meta: {
      title: 'TPPs',
      icon: Settings01Icon,
      description: 'Configure tags and contexts (colors, order).',
    }
  },
  {
    name: 'Settings',
    path: '/settings',
    component: Options,
    meta: {
      title: 'Settings',
      icon: Settings01Icon,
      description: 'Server and client settings.',
    }
  },
  {
    name: 'Diagnostics',
    path: '/diagnostics',
    component: Diagnostics,
    meta: {
      title: 'Diagnostics',
      icon: Settings01Icon,
      description: 'Debug and health information.',
    }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
