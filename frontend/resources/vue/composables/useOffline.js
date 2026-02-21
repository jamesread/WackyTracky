import { ref, onMounted, onUnmounted } from 'vue';

const isOnline = ref(typeof navigator !== 'undefined' ? navigator.onLine : true);

export function useOffline() {
  function updateOnline() {
    isOnline.value = navigator.onLine;
  }

  onMounted(() => {
    window.addEventListener('online', updateOnline);
    window.addEventListener('offline', updateOnline);
  });

  onUnmounted(() => {
    window.removeEventListener('online', updateOnline);
    window.removeEventListener('offline', updateOnline);
  });

  return { isOnline };
}

export { isOnline };
