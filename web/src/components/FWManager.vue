<template lang="pug">
  v-data-table(:headers="fwsHeaders" :items="fwsView" :items-per-page="-1")
    template(v-slot:top)
      v-toolbar(flat)
        v-toolbar-title Forwards
        v-spacer
        v-dialog(v-model="dialogVisible" max-width="32em")
          template(v-slot:activator="{ on }")
            v-btn(color="primary" v-on="on")
              v-icon mdi-plus-thick
          v-card
            v-card-title
              span.headline Add Forwarding rule
            v-card-text
              v-container
                v-row
                  v-col(cols="12" md="8")
                    v-select(v-model="editedFw.hypervisor" :items="agentNames" label="hypervisor")
                v-row
                  v-col(cols="12" md="8")
                    v-select(v-model="editedFw.proto" :items="protocols" label="protocol")
                v-row
                  v-col(cols="12" sm="6" md="4")
                    v-text-field(v-model="editedFw.from_port" label="from port")
                  v-col(cols="12" sm="6" md="4")
                    v-select(v-model="editedFw.to_name" :items="vmNames" label="to name")
                  v-col(cols="12" sm="6" md="4")
                    v-text-field(v-model="editedFw.to_port" label="to port")
                v-row
                  v-col(cols="12" md="8")
                    v-select(v-model="editedFw.type" :items="fwsType" label="type")
                v-row
                  v-col(cols="12")
                    v-text-field(v-model="editedFw.description" label="description")
            v-card-actions
              v-btn(color="darken-1" text @click="clear") Cancel
              v-spacer
              v-btn(color="blue darken-1" text @click="createFw") Create
    template(v-slot:item.link="{ item }")
      template(v-if="item.type === 'http/https'")
        a.link-text(:href="'http://' + getAgentIP(item.hypervisor) + ':' + item.from_port") http
        br
        a.link-text(:href="'https://' + getAgentIP(item.hypervisor) + ':' + item.from_port") https
      template(v-if="item.type === 'ssh'")
        a.link-text(:href="'ssh://' + getAgentIP(item.hypervisor) + ':' + item.from_port") ssh
      template(v-if="item.type === 'vnc'")
        a.link-text(:href="'vnc://' + getAgentIP(item.hypervisor) + ':' + item.from_port") vnc
    template(v-slot:item.action="{ item }")
      v-icon.mr-2(small @click="deleteItem(item)") mdi-delete
</template>

<script>
import util from "@/util";
import axios from "axios";
axios.defaults.headers.post["Content-Type"] = "application/json";

export default {
  name: "FWManager",
  props: ["agents", "fws", "vms"],
  data() {
    const fwHeaderList = ["proto", "translation", "description"];
    let fwHeaders;
    fwHeaders = fwHeaderList.map(x => ({ text: x, value: x }));
    fwHeaders.push({ text: "link", value: "link", sortable: false });
    fwHeaders.push({ text: "action", value: "action", sortable: false });

    const fwsType = ["http/https", "ssh", "vnc"];

    return {
      fwsHeaders: fwHeaders,
      fwsType: fwsType,
      dialogVisible: false,
      editedFw: { proto: "tcp" },
      protocols: ["tcp", "udp"]
    };
  },
  computed: {
    agentNames() {
      return this.agents.map(x => x.name);
    },
    vmNames() {
      return ["localhost"].concat(this.vms.map(x => x.name));
    },
    fwsView() {
      return this.fws.map(origFw => {
        var fw = Object.assign(origFw);
        fw.translation = `${fw.hypervisor}:${fw.from_port} -> ${fw.to_name}:${fw.to_port}`;
        return fw;
      });
    }
  },
  methods: {
    clear() {
      this.editedFw = { proto: "tcp" };
      this.dialogVisible = false;
    },

    createFw() {
      const ep = this.getAgentEndpoint(this.editedFw.hypervisor);
      if (ep === "") {
        alert("Unknown hypervisor.");
        this.clear();
        return;
      }
      const url = ep + "forwards";
      const body = this.editedFw;
      const errMsg = "Failed to create new forward";
      util
        .callAxios(axios.post, url, body, errMsg)
        .catch(msg => {
          this.$emit("push-toast", msg);
        })
        .finally(() => {
          this.clear();
          this.$emit("update-forwards");
        });
    },

    deleteItem(item) {
      if (confirm("Are you sure you want to delete this item?")) {
        console.log(item);
        const ep = this.getAgentEndpoint(item.hypervisor);
        const url = ep + "forwards";
        const body = { data: item };
        const errMsg = "Failed to delete forward";
        util
          .callAxios(axios.delete, url, body, errMsg)
          .catch(msg => {
            this.$emit("push-toast", msg);
          })
          .finally(() => {
            this.$emit("update-forwards");
          });
      }
    },

    getAgentEndpoint(name) {
      const target = this.agents.filter(x => x.name === name);
      if (target.length !== 1) {
        return "";
      }
      return target[0].api;
    },
    getAgentIP(name) {
      const target = this.agents.filter(x => x.name === name);
      if (target.length !== 1) {
        return "";
      }
      const url = new URL(target[0].api);
      return url.hostname;
    }
  }
};
</script>
