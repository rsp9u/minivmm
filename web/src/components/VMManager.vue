<template lang="pug">
  #minivmm_vmmanager.section
    div(style="display: flex")
      p.is-size-4 VMs
      b-button(type="is-info" icon-left="plus-thick" style="margin-left: auto" @click="dialogVisible = true") New
    b-table(:data="vms")
      template(v-slot:default="{ row: item }")
        b-table-column(v-for="attr in vmAttrs" :key="attr" :label="attr" :field="attr") {{ item[attr] }}
        b-table-column(label="lock")
          b-icon(size="is-small" :icon="item['lock'] === 'true' ? 'lock' : 'lock-open-outline'")
        b-table-column(label="action")
          VMMenu(:endpoint="getAgentEndpoint(item.hypervisor)" :item="item" @push-toast="propagatePushToast")
    b-modal(:active.sync="dialogVisible" width="32em")
      form
        .modal-card(style="max-width: 32em")
          header.modal-card-head
            p.modal-card-title Create VM
          section.modal-card-body
            .columns.is-multiline
              .column.is-6
                b-field(label="hypervisor")
                  b-select(v-model="editedVM.hypervisor" expanded @input="updateImages")
                    option(v-for="option in agentNames" :key="option" :value="option") {{ option }}
              .column.is-6
                b-field(label="image")
                  b-select(v-model="editedVM.image" expanded)
                    option(v-for="option in images" :key="option" :value="option") {{ option }}
              .column.is-12
                b-field(label="name")
                  b-input(v-model="editedVM.name")
              .column.is-4
                b-field(label="vcpu")
                  b-input(v-model="editedVM.cpu")
              .column.is-4
                b-field(label="memory")
                  b-input(v-model="editedVM.memory" placeholder="e.g. 521M 2048M")
              .column.is-4
                b-field(label="disk")
                  b-input(v-model="editedVM.disk" placeholder="e.g. 1024M 20G")
              .column.is-4
                b-field(label="add ssh forward")
                  b-checkbox(v-model="editedVM.ssh_fw") enable
              .column.is-8
                b-field(label="from port")
                  b-input(v-model="editedVM.ssh_fw_port" :disabled="!editedVM.ssh_fw")
              .column.is-12
                b-field(label="user data template")
                  b-select(v-model="editedVM.user_data_template" expanded @input="selectCloudinitTemplate")
                    option(v-for="option in cloudinitTemplates" :key="option.text" :value="option.value") {{ option.text }}
              .column.is-12
                b-field(label="user data")
                  b-input(v-model="editedVM.user_data" type="textarea")
          footer.modal-card-foot(style="justify-content: flex-end")
            b-button(@click="clear") Cancel
            b-button(type="is-info" @click="createVM") Create
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
    const menuItems = [
      { title: "start" },
      { title: "stop" },
      { title: "resize" },
      { title: "delete" }
    ];

    const defaultVM = {
      name: "",
      ssh_fw: true
    };

    return {
      vmAttrs: ["name", "status", "hypervisor", "image", "ip", "cpu", "memory", "disk"],
      dialogVisible: false,
      editedVM: Object.assign({}, defaultVM),
      defaultVM: defaultVM,
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
      this.editedVM = Object.assign({}, this.defaultVM);
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
          const successMsg = "Suceeded VM creation";
          this.$emit("push-toast", { message: successMsg, color: "is-success" });
          if (this.editedVM.ssh_fw) {
            const fw = {
              hypervisor: this.editedVM.hypervisor,
              proto: "tcp",
              from_port: this.editedVM.ssh_fw_port,
              to_name: this.editedVM.name,
              to_port: "22",
              description: `ssh to ${this.editedVM.name}`,
              type: "ssh"
            };
            this.$emit("add-forward", fw);
          }
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
