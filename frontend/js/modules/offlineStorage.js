/**
 * Offline task storage. Offline-only tasks and pending edits are stored in localStorage
 * and synced when the user comes back online.
 */

const OFFLINE_TASKS_KEY = 'wackytracky_offline_tasks';
const OFFLINE_EDITS_KEY = 'wackytracky_offline_edits';
const CACHED_INBOX_KEY = 'wackytracky_cached_inbox';
const CACHED_LISTS_KEY = 'wackytracky_cached_lists';

/** @typedef {{ id: string, content: string, parentId?: string, parentType?: string, tags?: string[], contexts?: string[], createdAt: string }} OfflineTask */
/** @typedef {{ content: string }} OfflineEdit */
/** @typedef {{ listId: string, listTitle: string, tasks: object[], tree: Record<string, { ids: string[] }>, hiddenTagNames: string[], hiddenContextNames: string[], timestamp: number }} CachedList */

export function getOfflineTasks() {
  try {
    const raw = localStorage.getItem(OFFLINE_TASKS_KEY);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

export function addOfflineTask(task) {
  const list = getOfflineTasks();
  list.push(task);
  try {
    localStorage.setItem(OFFLINE_TASKS_KEY, JSON.stringify(list));
  } catch (e) {
    console.warn('offlineStorage: failed to persist offline task', e);
  }
}

export function removeOfflineTask(id) {
  const list = getOfflineTasks().filter((t) => t.id !== id);
  try {
    localStorage.setItem(OFFLINE_TASKS_KEY, JSON.stringify(list));
  } catch (e) {
    console.warn('offlineStorage: failed to remove offline task', e);
  }
}

export function clearOfflineTasks() {
  try {
    localStorage.removeItem(OFFLINE_TASKS_KEY);
  } catch {}
}

export function getOfflineEdits() {
  try {
    const raw = localStorage.getItem(OFFLINE_EDITS_KEY);
    return raw ? JSON.parse(raw) : {};
  } catch {
    return {};
  }
}

export function setOfflineEdit(taskId, content) {
  const edits = getOfflineEdits();
  edits[taskId] = { content };
  try {
    localStorage.setItem(OFFLINE_EDITS_KEY, JSON.stringify(edits));
  } catch (e) {
    console.warn('offlineStorage: failed to persist offline edit', e);
  }
}

export function removeOfflineEdit(taskId) {
  const edits = getOfflineEdits();
  delete edits[taskId];
  try {
    localStorage.setItem(OFFLINE_EDITS_KEY, JSON.stringify(edits));
  } catch (e) {
    console.warn('offlineStorage: failed to remove offline edit', e);
  }
}

export function clearOfflineEdits() {
  try {
    localStorage.removeItem(OFFLINE_EDITS_KEY);
  } catch {}
}

export function getCachedInbox() {
  try {
    const raw = localStorage.getItem(CACHED_INBOX_KEY);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

export function setCachedInbox(data) {
  try {
    localStorage.setItem(CACHED_INBOX_KEY, JSON.stringify({
      ...data,
      timestamp: Date.now(),
    }));
    const lists = getCachedLists();
    lists[data.listId] = { ...data, timestamp: Date.now() };
    localStorage.setItem(CACHED_LISTS_KEY, JSON.stringify(lists));
  } catch (e) {
    console.warn('offlineStorage: failed to cache inbox', e);
  }
}

export function clearCachedInbox() {
  try {
    localStorage.removeItem(CACHED_INBOX_KEY);
  } catch {}
}

export function getCachedList(listId) {
  try {
    const raw = localStorage.getItem(CACHED_LISTS_KEY);
    const lists = raw ? JSON.parse(raw) : {};
    return lists[listId] || null;
  } catch {
    return null;
  }
}

export function getCachedLists() {
  try {
    const raw = localStorage.getItem(CACHED_LISTS_KEY);
    return raw ? JSON.parse(raw) : {};
  } catch {
    return {};
  }
}

export function generateOfflineTaskId() {
  return `offline-${Date.now()}-${Math.random().toString(36).slice(2, 11)}`;
}

export const INBOX_LIST_ID = 'inbox';
