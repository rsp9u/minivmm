<template lang="pug">
  #minivmm_fwmanager.section
    div(style="display: flex")
      p.is-size-4 Forwards
      b-button(type="is-info" icon-left="plus-thick" style="margin-left: auto" @click="dialogVisible = true") New
    b-table(:data="fwsView" :hoverable="true")
      template(v-slot:default="{ row: item }")
        b-table-column(v-for="attr in fwAttrs" :key="attr" :label="attr" :field="attr") {{ item[attr] }}
        b-table-column(label="link")
          template(v-if="item.type === 'http/https'")
            a.link-text(:href="'http://' + getAgentIP(item.hypervisor) + ':' + item.from_port") http
            br
            a.link-text(:href="'https://' + getAgentIP(item.hypervisor) + ':' + item.from_port") https
          template(v-if="item.type === 'ssh'")
            a.link-text(:href="'ssh://' + getAgentIP(item.hypervisor) + ':' + item.from_port") ssh
          template(v-if="item.type === 'vnc'")
            a.link-text(:href="'vnc://' + getAgentIP(item.hypervisor) + ':' + item.from_port") vnc
        b-table-column(label="action")
          template(slot="header" slot-scope="{ column }")
            span.tag.is-danger {{ column.label }}
          b-tooltip(label="delete" position="is-right")
            b-button(type="is-text" icon-left="delete" size="is-small" @click="deleteItem(item)")
    b-modal(:active.sync="dialogVisible" width="32em" :can-cancel="['escape', 'outside']")
      FWDialog(ref="dialog" :agents="agents" :vms="vms" @create="createFw" @cancel="dialogVisible = false;")
</template>

<script>
import util from "@/util";
import axios from "axios";
import FWDialog from "@/components/FWDialog";
axios.defaults.headers.post["Content-Type"] = "application/json";

export default {
  name: "FWManager",
  components: {FWDialog},
  props: ["agents", "fws", "vms"],
  data() {
    return {
      fwAttrs: ["proto", "translation", "description"],
      dialogVisible: false,
    };
  },
  computed: {
    fwsView() {
      return this.fws.map(origFw => {
        var fw = Object.assign(origFw);
        fw.translation = `${fw.hypervisor}:${fw.from_port} -> ${fw.to_name}:${fw.to_port}`;
        return fw;
      });
    }
  },
  methods: {
    clearDialog() {
      this.dialogVisible = false;
      this.$refs.dialog.clear();
    },

    createFw(fw) {
      const ep = this.getAgentEndpoint(fw.hypervisor);
      if (ep === "") {
        alert("Unknown hypervisor.");
        this.clearDialog();
        return;
      }
      const url = ep + "forwards";
      const body = fw;
      const errMsg = "Failed to create new forward";
      util
        .callAxios(axios.post, url, body, errMsg)
        .catch(msg => {
          this.$emit("push-toast", msg);
        })
        .finally(() => {
          this.clearDialog();
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
