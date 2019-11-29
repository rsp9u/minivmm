<template lang="pug">
  div#minivmm_main
    v-app
      v-content
        v-container
          VMManager(:agents="agents" :vms="vms" @push-toast="toast" @update-vms="getAllVMs")
          FWManager(:agents="agents" :fws="fws" :vms="vms" @push-toast="toast" @update-forwards="getAllForwards")
      v-snackbar(
        v-model="toastVisible"
        :top="true"
        :multi-line="true"
        :color="toastColor"
      ) {{ toastMessage }}
</template>

<script>
import util from "@/util";
import axios from "axios";

import VMManager from "@/components/VMManager";
import FWManager from "@/components/FWManager";

export default {
  name: "minivm-main",
  components: {
    VMManager,
    FWManager
  },
  data() {
    return {
      agents: [],
      vms: [],
      fws: [],
      intervalIds: [],
      toastMessage: "",
      toastColor: "",
      toastVisible: false
    };
  },
  created() {
    this.getAgents().then(() => {
      this.getAllVMs();
      this.getAllForwards();
    });
    this.setPoll();
  },
  beforeDestroy() {
    this.clearPoll();
  },
  methods: {
    // Agent
    async getAgents() {
      util.ensureAxiosAuth();
      const response = await axios.get(
        util.locationOrigin() + "/api/v1/agents"
      );
      this.agents = response.data.agents;
    },
    // VM
    async getVMs(apiEndpoint) {
      util.ensureAxiosAuth();
      try {
        const response = await axios.get(apiEndpoint + "vms");
        const ipUpdated = response.data.vms.map(vm =>
          vm.ip === "" ? Object.assign(vm, { ip: "requesting.." }) : vm
        );
        return ipUpdated;
      } catch {
        return [];
      }
    },
    async getAllVMs() {
      let vms = [];
      for (let agent of this.agents) {
        const ret = await this.getVMs(agent.api);
        vms.push(...ret);
      }
      if (this.diffVMs(this.vms, vms)) {
        this.vms = vms;
      }
    },
    diffVMs(prevVMs, currVMs) {
      prevVMs.sort();
      currVMs.sort();
      return JSON.stringify(prevVMs) !== JSON.stringify(currVMs);
    },
    // Forward
    async getForwards(apiEndpoint) {
      util.ensureAxiosAuth();
      try {
        const response = await axios.get(apiEndpoint + "forwards");
        return response.data.forwards;
      } catch {
        return [];
      }
    },
    async getAllForwards() {
      let fws = [];
      for (let agent of this.agents) {
        const ret = await this.getForwards(agent.api);
        fws.push(...ret);
      }
      if (this.diffForwards(this.fws, fws)) {
        this.fws = fws;
      }
    },
    diffForwards(prevFws, currFws) {
      prevFws.sort();
      currFws.sort();
      return JSON.stringify(prevFws) !== JSON.stringify(currFws);
    },
    // Polling
    setPoll() {
      this.intervalIds.push(setInterval(this.getAllVMs, 5000));
      this.intervalIds.push(setInterval(this.getAllForwards, 5000));
    },
    clearPoll() {
      console.log("clearInterval");
      for (var id of this.intervalIds) {
        clearInterval(id);
      }
    },
    // Toast
    toast({ message, color }) {
      this.toastMessage = message;
      this.toastColor = color;
      this.toastVisible = true;
    }
  }
};
</script>
