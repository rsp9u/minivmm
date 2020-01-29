<template lang="pug">
  #minivmm_fwmanager.section
    div(style="display: flex")
      p.is-size-4 Forwards
      b-button(type="is-info" icon-left="plus-thick" style="margin-left: auto" @click="dialogVisible = true") New
    b-table(:data="fwsView")
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
          b-tooltip(label="delete" position="is-right")
            b-button(type="is-text" icon-left="delete" size="is-small" @click="deleteItem(item)")
    b-modal(:active.sync="dialogVisible" width="32em" :can-cancel="['escape', 'outside']")
      form
        .modal-card(style="max-width: 32em")
          header.modal-card-head
            p.modal-card-title Add Forwarding rule
          section.modal-card-body
            .columns.is-multiline
              .column.is-12
                b-field(label="hypervisor")
                  b-select(v-model="editedFw.hypervisor" expanded)
                    option(v-for="option in agentNames" :key="option" :value="option") {{ option }}
              .column.is-3
                b-field(label="protocol")
                  b-select(v-model="editedFw.proto" expanded)
                    option(v-for="option in protocols" :key="option" :value="option") {{ option }}
              .column.is-3
                b-field(label="from port")
                  b-input(v-model="editedFw.from_port")
              .column.is-3
                b-field(label="to name")
                  b-select(v-model="editedFw.to_name" expanded)
                    option(v-for="option in vmNames" :key="option" :value="option") {{ option }}
              .column.is-3
                b-field(label="to port")
                  b-input(v-model="editedFw.to_port")
              .column.is-6
                b-field(label="description")
                  b-input(v-model="editedFw.description")
              .column.is-6
                b-field(label="type")
                  b-select(v-model="editedFw.type" expanded)
                    option(v-for="option in fwTypes" :key="option" :value="option") {{ option }}
          footer.modal-card-foot(style="justify-content: flex-end")
            b-button(@click="clear") Cancel
            b-button(type="is-info" @click="createFw") Create
</template>

<script>
import util from "@/util";
import axios from "axios";
axios.defaults.headers.post["Content-Type"] = "application/json";

export default {
  name: "FWManager",
  props: ["agents", "fws", "vms"],
  data() {
    return {
      fwAttrs: ["proto", "translation", "description"],
      fwTypes: ["http/https", "ssh", "vnc"],
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
      this.createFwFromObject(this.editedFw);
    },
    createFwFromObject(fw) {
      const ep = this.getAgentEndpoint(fw.hypervisor);
      if (ep === "") {
        alert("Unknown hypervisor.");
        this.clear();
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
