import c from "@/const";
import Cookies from "js-cookie";
import axios from "axios";

function ensureAxiosAuth() {
  const token = Cookies.get(c.COOKIE_TOKEN);
  if (token !== undefined) {
    axios.defaults.headers.common["Authorization"] = `Bearer ${token}`;
  }
}

function callAxios(axiosFunc, url, body, errMsg) {
  ensureAxiosAuth();
  if (body === null) {
    return axiosFunc(url).catch(error => {
      const body = error.response.data;
      const msg = `${errMsg}: ${body.error}`;
      throw { message: msg, color: "error" };
    });
  } else {
    return axiosFunc(url, body).catch(error => {
      const body = error.response.data;
      const msg = `${errMsg}: ${body.error}`;
      throw { message: msg, color: "error" };
    });
  }
}

function locationOrigin() {
  if (process.env.VUE_APP_LOCATION_ORIGIN !== undefined) {
    return process.env.VUE_APP_LOCATION_ORIGIN;
  } else {
    return window.location.origin;
  }
}

export default {
  ensureAxiosAuth,
  callAxios,
  locationOrigin
};
