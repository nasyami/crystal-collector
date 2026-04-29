
export const GTAG_EVENTS = {
  VIEW_ITEM_LIST: 'shop_main_list_view',
  BEGIN_CHECKOUT: 'shop_item_buy_click',
  LOGIN: 'login_auth_success',
  PURCHASE: 'shop_checkout_payment_succeeded',
  PAYMENT_FAILED: 'shop_checkout_payment_failed',
} as const;

export type GtagEventName = typeof GTAG_EVENTS[keyof typeof GTAG_EVENTS];

/**
 * Global wrapper to ensure typo-proof event firing
 */
export function fireEvent(eventName: GtagEventName, params?: Record<string, unknown>): void {
  if (typeof gtag !== 'function') return;
  gtag('event', eventName, params);
}
