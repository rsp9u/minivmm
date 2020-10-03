<template lang="pug">
  #minivmm_main
    VMManager(ref="vmm" :agents="agents" :vms="vms" @push-toast="toast" @update-vms="getAllVMs" @add-forward="addForwardFromVMManager" @delete-vm="deleteForwardByVMDeletion")
    FWManager(ref="fwm" :agents="agents" :fws="fws" :vms="vms" @push-toast="toast" @update-forwards="getAllForwards")
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
      intervalIds: []
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
      const response = await axios.get(
        util.locationOrigin() + "/api/v1/agents"
      );
      this.agents = response.data.agents;
    },
    // VM
    async getVMs(apiEndpoint) {
      try {
        const response = await axios.get(apiEndpoint + "/vms");
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
    addForwardFromVMManager(fw) {
      this.$refs.fwm.createFw(fw);
    },
    deleteForwardByVMDeletion(deletedVM) {
      this.$refs.fwm.deleteItemsRelatedVM(deletedVM.hypervisor, deletedVM.name);
    },
    // Forward
    async getForwards(apiEndpoint) {
      try {
        const response = await axios.get(apiEndpoint + "/forwards");
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
    toast({ message, color, duration }) {
      if (duration) {
        this.$buefy.toast.open({ message: message, type: color, duration: duration })
      } else {
        this.$buefy.toast.open({ message: message, type: color })
      }
    }
  }
};
</script>
