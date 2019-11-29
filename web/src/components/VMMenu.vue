<template lang="pug">
  #vmmenu
    v-menu(offset-y)
      template(v-slot:activator="{ on }")
        v-icon.mr-2(v-on="on") mdi-dots-horizontal
      v-list
        v-list-item(v-for="(menu, menuIndex) in menuItems" :key="menuIndex" @click="clickMenu(menu)")
          v-list-item-title {{ menu.title }}
    v-dialog(v-model="resizeDialog" persistent max-width="30em")
      v-card
        v-card-title.headline Resize VM
        v-card-text
          v-container
            v-row
              v-col(cols="12")
                v-select(v-model="editedResize.type" :items="resizeType" label="type")
            template(v-if="editedResize.type === 'cpu/memory'")
              v-row
                v-col(cols="6")
                  v-text-field(v-model="editedResize.cpu" label="vcpu")
                v-col(cols="6")
                  v-text-field(v-model="editedResize.memory" label="memory" placeholder="e.g. 521M 2048M")
            template(v-if="editedResize.type === 'disk'")
              v-row
                v-col(cols="6")
                  v-text-field(v-model="editedResize.disk" label="disk" placeholder="e.g. 1024M 20G")
        v-card-actions
          v-btn(color="darken-1" text @click="resizeCancel") Cancel
          v-spacer
          v-btn(color="blue darken-1" text @click="resizeVM") Create
</template>

<script>
import util from "@/util";
import axios from "axios";
axios.defaults.headers.post["Content-Type"] = "application/json";

export default {
  name: "VMMenu",
  props: ["endpoint", "item"],
  data() {
    const menuItems = [
      { title: "start" },
      { title: "stop" },
      { title: "resize" },
      { title: "delete" }
    ];

    return {
      menuItems: menuItems,
      editedResize: {},
      resizeDialog: false,
      resizeType: ["cpu/memory", "disk"]
    };
  },
  methods: {
    clickMenu(menu) {
      switch (menu.title) {
        case "start":
          this.startVM();
          break;
        case "stop":
          this.stopVM();
          break;
        case "resize":
          this.resizeDialog = true;
          break;
        case "delete":
          this.deleteVM();
          break;
        default:
          console.log("This menu hasn't implemented yet: ", menu);
      }
    },
    startVM() {
      console.log(this.item);
      this.updateVMStatus(this.item.name, "start");
    },
    stopVM() {
      console.log(this.item);
      this.updateVMStatus(this.item.name, "stop");
    },
    updateVMStatus(name, status) {
      const url = this.endpoint + `vms/${name}`;
      const body = { status: status };
      const errMsg = "Failed to change VM status";
      util
        .callAxios(axios.patch, url, body, errMsg)
        .catch(msg => {
          this.$emit("push-toast", msg);
        })
        .finally(() => {
          this.$emit("update-vms");
        });
    },
    resizeVM() {
      console.log(this.item);
      const url = this.endpoint + `vms/${this.item.name}`;
      const body = this.editedResize;
      const errMsg = "Failed to resize VM";
      util
        .callAxios(axios.patch, url, body, errMsg)
        .catch(msg => {
          this.$emit("push-toast", msg);
        })
        .finally(() => {
          this.resizeDialog = false;
          this.editedResize = {};
          this.$emit("update-vms");
        });
    },
    resizeCancel() {
      this.resizeDialog = false;
      this.editedResize = {};
    },
    deleteVM() {
      if (confirm("Are you sure you want to delete this item?")) {
        const infoMsg = "Accepted VM deletion";
        this.$emit("push-toast", { message: infoMsg, color: "info" });

        const url = this.endpoint + `vms/${this.item.name}`;
        const errMsg = "Failed to delete VM";
        util
          .callAxios(axios.delete, url, null, errMsg)
          .then(() => {
            const successMsg = "Suceeded VM deletion";
            this.$emit("push-toast", { message: successMsg, color: "success" });
          })
          .catch(msg => {
            this.$emit("push-toast", msg);
          })
          .finally(() => {
            this.$emit("update-vms");
          });
      }
    }
  }
};
</script>
