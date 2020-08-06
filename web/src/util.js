function callAxios(axiosFunc, url, body, errMsg) {
  let opts =  {withCredentials: true};
  if (body === null) {
    return axiosFunc(url, opts).catch(error => {
      const body = error.response.data;
      const msg = `${errMsg}: ${body.error}`;
      throw { message: msg, color: "is-danger", duration: 5000 };
    });
  } else {
    return axiosFunc(url, body, opts).catch(error => {
      const body = error.response.data;
      const msg = `${errMsg}: ${body.error}`;
      throw { message: msg, color: "is-danger", duration: 5000 };
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
  callAxios,
  locationOrigin
};
