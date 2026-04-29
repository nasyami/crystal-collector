declare function gtag(...args: unknown[]): void;

export const GTAG_EVENTS = {
  VIEW_ITEM: 'view_item',
  ADD_TO_CART: 'add_to_cart',
  BEGIN_CHECKOUT: 'begin_checkout',
  PURCHASE: 'purchase',
  PAYMENT_FAILED: 'payment_failed', 
} as const;