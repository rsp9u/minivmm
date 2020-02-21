<template lang="pug">
  #minivmm_vmmenu
    b-dropdown(aria-role="list")
      template(v-slot:trigger)
        b-button(type="is-text" icon-left="dots-horizontal" size="is-small")
      b-dropdown-item(
        v-for="(menu, menuIndex) in menuItems"
        :key="menuIndex"
        aria-role="listitem"
        :disabled="menu.title === 'delete' && item.lock === 'true'"
        @click="clickMenu(menu)"
      ) {{ menu.title }}
    b-modal(:active.sync="resizeDialog" width="30em" :can-cancel="['escape', 'outside']")
      .modal-card(style="max-width: 30em")
        header.modal-card-head
          p.modal-card-title Resize VM
        section.modal-card-body
          .columns.is-multiline
            .column.is-12
              b-field(label="type")
                b-select(v-model="editedResize.type" expanded)
                  option(v-for="option in resizeType" :key="option" :value="option") {{ option }}
            template(v-if="editedResize.type === 'cpu/memory'")
              .column.is-6
                b-field(label="vcpu")
                  b-input(v-model="editedResize.cpu")
              .column.is-6
                b-field(label="memory")
                  b-input(v-model="editedResize.memory" placeholder="e.g. 521M 2048M")
            template(v-if="editedResize.type === 'disk'")
              .column.is-6
                b-field(label="disk")
                  b-input(v-model="editedResize.disk" placeholder="e.g. 1024M 20G")
        footer.modal-card-foot(style="justify-content: flex-end")
          b-button(@click="resizeCancel") Cancel
          b-button(type="is-info" @click="resizeVM") Resize
</template>

<script>
import util from "@/util";
import axios from "axios";
axios.defaults.headers.post["Content-Type"] = "application/json";

export default {
  name: "VMMenu",
  props: ["endpoint", "item"],
  data() {

    return {
      editedResize: {},
      resizeDialog: false,
      resizeType: ["cpu/memory", "disk"]
    };
  },
  computed: {
    menuItems() {
      let menuItems = [
        { title: "start" },
        { title: "stop" },
        { title: "resize" },
        { title: "lock/unlock" },
        { title: "delete" }
      ];
      if (this.item.lock === "true") {
        menuItems[3].title = "unlock";
      } else {
        menuItems[3].title = "lock";
      }
      return menuItems;
    }
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
        case "lock":
          this.setLockVM("true");
          break;
        case "unlock":
          this.setLockVM("false");
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
    setLockVM(lock) {
      console.log(this.item);
      const lockOrUnlock = lock === "true" ? "lock" : "unlock";
      const infoMsg = `Accepted VM ${lockOrUnlock}`;
      this.$emit("push-toast", { message: infoMsg, color: "is-info" });

      const url = this.endpoint + `vms/${this.item.name}`;
      const body = {lock: lock};
      const errMsg = `Failed to ${lockOrUnlock} VM`;
      util
        .callAxios(axios.patch, url, body, errMsg)
        .catch(msg => {
          this.$emit("push-toast", msg);
        })
        .finally(() => {
          this.$emit("update-vms");
        });
    },
    deleteVM() {
      if (confirm("Are you sure you want to delete this item?")) {
        const infoMsg = "Accepted VM deletion";
        this.$emit("push-toast", { message: infoMsg, color: "is-info" });

        const url = this.endpoint + `vms/${this.item.name}`;
        const errMsg = "Failed to delete VM";
        util
          .callAxios(axios.delete, url, null, errMsg)
          .then(() => {
            const successMsg = "Suceeded VM deletion";
            this.$emit("push-toast", { message: successMsg, color: "is-success" });
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
