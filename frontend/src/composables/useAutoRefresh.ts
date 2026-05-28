import { onBeforeUnmount, onMounted } from "vue";

/**
 * Запускает переданный refresh-колбэк, когда пользователь возвращается в вкладку
 * (visibilitychange) или окно получает фокус. Это нужно, чтобы админские страницы
 * подхватывали изменения, сделанные другими администраторами, без необходимости
 * полностью перезагружать приложение.
 *
 * Колбэк не вызывается на mount — для первичной загрузки используйте обычный onMounted
 * с `force: true`.
 */
export function useAutoRefresh(refresh: () => void | Promise<void>) {
  let isRunning = false;

  const trigger = async () => {
    if (document.visibilityState !== "visible") return;
    if (isRunning) return;
    isRunning = true;
    try {
      await refresh();
    } catch {
      // Сетевые ошибки логируются внутри сторов; здесь они нам не интересны.
    } finally {
      isRunning = false;
    }
  };

  onMounted(() => {
    document.addEventListener("visibilitychange", trigger);
    window.addEventListener("focus", trigger);
  });

  onBeforeUnmount(() => {
    document.removeEventListener("visibilitychange", trigger);
    window.removeEventListener("focus", trigger);
  });
}
