function callAxios(axiosFunc, url, body, errMsg) {
  if (body === null) {
    return axiosFunc(url).catch(error => {
      const body = error.response.data;
      const msg = `${errMsg}: ${body.error}`;
      throw { message: msg, color: "is-danger", duration: 5000 };
    });
  } else {
    return axiosFunc(url, body).catch(error => {
      const body = error.response.data;
      const msg = `${errMsg}: ${body.error}`;
      throw { message: msg, color: "is-danger", duration: 5000 };
    });
  }
}

function trimTailingSlash(url) {
  return url.replace(/\/+$/, '');
}

function locationOrigin() {
  if (process.env.VUE_APP_LOCATION_ORIGIN !== undefined) {
    return process.env.VUE_APP_LOCATION_ORIGIN;
  } else {
    return window.location.origin;
  }
}

export default {
  callAxios,
  trimTailingSlash,
  locationOrigin
};
