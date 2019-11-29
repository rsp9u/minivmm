import Vue from "vue";
import VueRouter from "vue-router";
import Home from "../views/Home.vue";
import Login from "../views/Login.vue";
import util from "../util";

import axios from "axios";

Vue.use(VueRouter);

const routes = [
  {
    path: "/",
    name: "home",
    component: Home,
    meta: { requiresAuth: true }
  },
  {
    path: "/login",
    name: "login",
    component: Login
  }
];

const router = new VueRouter({
  mode: "history",
  base: process.env.BASE_URL,
  routes
});

router.beforeEach((to, from, next) => {
  if (!to.matched.some(record => record.meta.requiresAuth)) {
    next();
    return;
  }

  util.ensureAxiosAuth();
  axios
    .get(util.locationOrigin() + "/api/v1/auth")
    .then(response => {
      if (response.status !== 200) {
        next({ path: "/login" });
      } else {
        next();
      }
    })
    .catch(() => {
      next({ path: "/login" });
    });
});

export default router;
