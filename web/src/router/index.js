import Vue from "vue";
import VueRouter from "vue-router";
import Main from "../components/Main.vue";
import NoVNC from "../components/NoVNC.vue";

Vue.use(VueRouter);

const routes = [
  {
    path: "/",
    name: "main",
    component: Main
  },
  {
    path: "/vnc",
    name: "vnc",
    component: NoVNC
  }
];

const router = new VueRouter({
  mode: "history",
  base: process.env.BASE_URL,
  routes
});

export default router;
