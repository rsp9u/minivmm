<template lang="pug">
  #vmmanager
    v-data-table(:headers="vmsHeaders" :items="vms")
      template(v-slot:top)
        v-toolbar(flat)
          v-toolbar-title VMs
          v-spacer
          v-dialog(v-model="dialogVisible" max-width="40em")
            template(v-slot:activator="{ on }")
              v-btn(color="primary" v-on="on")
                v-icon mdi-plus-thick
            v-card
              v-card-title
                span.headline Create VM
              v-card-text
                v-container
                  v-row
                    v-col(cols="12")
                      v-text-field(v-model="editedVM.name" :rules="[rules.max]" @update:error="nameError" label="name")
                    v-col(cols="12")
                      v-select(v-model="editedVM.hypervisor" :items="agentNames" label="hypervisor" @change="updateImages")
                    v-col(cols="12")
                      v-select(v-model="editedVM.image" :items="images" label="image")
                    v-col(cols="12" md="4")
                      v-text-field(v-model="editedVM.cpu" label="vcpu")
                    v-col(cols="12" md="4")
                      v-text-field(v-model="editedVM.memory" label="memory" placeholder="e.g. 521M 2048M")
                    v-col(cols="12" md="4")
                      v-text-field(v-model="editedVM.disk" label="disk" placeholder="e.g. 1024M 20G")
                    v-col(cols="12")
                      v-select(v-model="editedVM.user_data_template" :items="cloudinitTemplates" label="user data template" @change="selectCloudinitTemplate")
                    v-col(cols="12")
                      v-textarea(v-model="editedVM.user_data" label="user data")
              v-card-actions
                v-btn(color="darken-1" text @click="clear") Cancel
                v-spacer
                v-btn(color="blue darken-1" text :disabled="invalidVM" @click="createVM") Create
      template(v-slot:item.action="{ item }")
        VMMenu(:endpoint="getAgentEndpoint(item.hypervisor)" :item="item" @push-toast="propagatePushToast")
    v-dialog(v-model="vncPopup" persistent max-width="20em")
      v-card
        v-card-title.headline VNC Info
        v-card-text
          | Port: {{ vncPort }}
          br
          | Password: {{ vncPassword }}
        v-btn(color="green darken-1" text @click="vncPopup = false") Close
</template>

<script>
import util from "@/util";
import cloudinit from "@/cloudinit";
import axios from "axios";
axios.defaults.headers.post["Content-Type"] = "application/json";

import VMMenu from "@/components/VMMenu";

export default {
  name: "VMManager",
  components: {
    VMMenu
  },
  props: ["agents", "vms"],
  data() {
    const vmHeaderList = [
      "name",
      "status",
      "hypervisor",
      "ip",
      "cpu",
      "memory",
      "disk",
    ];
    let vmHeaders;
    vmHeaders = vmHeaderList.map(x => ({ text: x, value: x }));
    vmHeaders.push({ text: "action", value: "action", sortable: false });

    const menuItems = [
      { title: "start" },
      { title: "stop" },
      { title: "resize" },
      { title: "delete" }
    ];

    return {
      vmsHeaders: vmHeaders,
      dialogVisible: false,
      vncPort: "",
      vncPassword: "",
      vncPopup: false,
      editedVM: { name: "" },
      images: [],
      menuItems: menuItems,
      cloudinitTemplates: cloudinit.templates,
      rules: {
        max: v => v.length <= 10 || "Max 10 characters"
      },
      invalidVM: false
    };
  },
  computed: {
    agentNames() {
      return this.agents.map(x => x.name);
    }
  },
  methods: {
    clear() {
      this.editedVM = { name: "" };
      this.images = [];
      this.dialogVisible = false;
    },
    nameError(errorStatus) {
      this.invalidVM = errorStatus;
    },

    updateImages() {
      if (this.editedVM.hypervisor === "") {
        this.images = [];
        return;
      }
      const ep = this.getAgentEndpoint(this.editedVM.hypervisor);
      if (ep !== "") {
        axios
          .get(ep + "images")
          .then(
            response => (this.images = response.data.images.map(x => x.name))
          );
      }
    },

    selectCloudinitTemplate() {
      if (this.editedVM.user_data_template !== null) {
        this.editedVM.user_data = this.editedVM.user_data_template;
      }
    },

    createVM() {
      const ep = this.getAgentEndpoint(this.editedVM.hypervisor);
      if (ep === "") {
        alert("Unknown hypervisor.");
        this.clear();
        return;
      }

      const url = ep + "vms";
      const body = this.editedVM;
      const errMsg = "Failed to create new VM";
      util
        .callAxios(axios.post, url, body, errMsg)
        .then(response => {
          console.log("then");
          this.vncPort = response.data.vnc_port;
          this.vncPassword = response.data.vnc_password;
          this.vncPopup = true;
        })
        .catch(msg => {
          this.$emit("push-toast", msg);
        })
        .finally(() => {
          console.log("finally");
          this.clear();
          this.$emit("update-vms");
        });
    },

    getAgentEndpoint(name) {
      const target = this.agents.filter(x => x.name === name);
      if (target.length !== 1) {
        return "";
      }
      return target[0].api;
    },

    propagatePushToast(event) {
      this.$emit("push-toast", event);
    }
  }
};
</script>
