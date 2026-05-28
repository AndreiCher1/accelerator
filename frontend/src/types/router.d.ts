import "vue-router";

declare module "vue-router" {
  interface RouteMeta {
    title?: string;
    requiresAuth?: boolean;
    guestOnly?: boolean;
    requiresPasswordChange?: boolean;
    requiresAdmin?: boolean;
  }
}
