<template lang="pug">
  #minivmm_novnc
    .novnc-titlebar
      .novnc-titlebar-text
        b VNC Connection to "{{ name }}"
    #novnc
</template>

<script>
import util from "@/util";

import RFB from '@novnc/novnc/core/rfb';

export default {
  name: "minivm-novnc",
  data() {
    return {
      name: this.$route.query.name,
      rfb: null
    };
  },
  mounted() {
    this.$nextTick(() => {
      let el = document.getElementById("novnc");
      let url = `${util.locationOrigin()}/ws/vnc?name=${this.$route.query.name}`.replace(/^http/, "ws");
      let opts = {wsProtocols: ["binary"]};
      this.rfb = new RFB(el, url, opts);
    });
  }
}
</script>

<style>
.novnc-titlebar {
  background-color: #282828;
  outline-color: #4f4f4f;
  padding: 1em;
}

.novnc-titlebar-text {
  color: #00a040;
  text-align: center;
}

#minivmm_novnc {
  height: 100%;
  background-color: #282828;
}
</style>
