<template lang="pug">
  #minivmm_vmmenu
    b-dropdown(aria-role="list")
      template(v-slot:trigger)
        b-button(type="is-text" icon-left="dots-horizontal" size="is-small")
      b-dropdown-item(v-for="(menu, menuIndex) in menuItems" :key="menuIndex" aria-role="listitem" :disabled="menu.disabled" @click="clickMenu(menu)") {{ menu.title }}

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

    b-modal(:active.sync="addVolumeDialog" width="30em" :can-cancel="['escape', 'outside']")
      .modal-card(style="max-width: 30em")
        header.modal-card-head
          p.modal-card-title Add Volume
        section.modal-card-body
          .columns.is-multiline
            .column.is-12
              b-field(label="size")
                b-input(v-model="addVolumeSize")
        footer.modal-card-foot(style="justify-content: flex-end")
          b-button(@click="addVolumeCancel") Cancel
          b-button(type="is-info" @click="addVolume") Create

    b-modal(:active.sync="rmVolumeDialog" width="30em" :can-cancel="['escape', 'outside']")
      .modal-card(style="max-width: 30em")
        header.modal-card-head
          p.modal-card-title Remove Volume
        section.modal-card-body
          b-table(:data="item.extra_volumes" :narrowed="true" :hoverable="true" @click="rmVolume")
            template(v-slot:default="{ row: volume }")
              b-table-column(v-for="attr in ['name', 'size']" :key="attr" :label="attr" :field="attr") {{ volume[attr] }}
        footer.modal-card-foot(style="justify-content: flex-end")
          b-button(@click="rmVolumeCancel") Cancel
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
      resizeType: ["cpu/memory", "disk"],
      addVolumeDialog: false,
      addVolumeSize: "",
      rmVolumeDialog: false
    };
  },
  computed: {
    menuItems() {
      let items = [
        {
          title: "start",
          disabled: this.item.status === "running"
        },
        {
          title: "stop",
          disabled: this.item.status === "stopped"
        },
        {
          title: "vnc",
          disabled: this.item.status !== "running"
        },
        {
          title: "resize",
          disabled: this.item.status !== "stopped"
        },
        {
          title: "add volume",
          disabled: this.item.status === "running"
        },
        {
          title: "rm volume",
          disabled: this.item.status === "running" || this.item.lock === "true"
        },
        {
          title: this.item.lock === "true" ? "unlock" : "lock",
          disabled: false
        },
        {
          title: "delete",
          disabled: this.item.lock === "true"
        }
      ];
      return items;
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
        case "vnc":
          this.openVNC();
          break;
        case "resize":
          this.resizeDialog = true;
          break;
        case "add volume":
          this.addVolumeDialog = true;
          break;
        case "rm volume":
          this.rmVolumeDialog = true;
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
    openVNC() {
      let route = this.$router.resolve({ name: "vnc", query: { name: this.item.name } });
      window.open(route.href, "_blank");
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
    addVolume() {
      console.log(this.item);
      const url = this.endpoint + `vms/${this.item.name}/volumes`;
      const body = { size: this.addVolumeSize };
      const errMsg = "Failed to add a new volume";
      util
        .callAxios(axios.post, url, body, errMsg)
        .catch(msg => {
          this.$emit("push-toast", msg);
        })
        .finally(() => {
          this.addVolumeDialog = false;
          this.addVolumeSize = "";
          this.$emit("update-vms");
        });
    },
    rmVolume(volume) {
      console.log(volume);
      const url = this.endpoint + `vms/${this.item.name}/volumes/${volume.name}`;
      const errMsg = "Failed to remove a volume";
      util
        .callAxios(axios.delete, url, {}, errMsg)
        .catch(msg => {
          this.$emit("push-toast", msg);
        })
        .finally(() => {
          this.rmVolumeDialog = false;
          this.$emit("update-vms");
        });
    },
    resizeCancel() {
      this.resizeDialog = false;
      this.editedResize = {};
    },
    addVolumeCancel() {
      this.addVolumeDialog = false;
      this.addVolumeSize = "";
    },
    rmVolumeCancel() {
      this.rmVolumeDialog = false;
    },
    setLockVM(lock) {
      console.log(this.item);
      const lockOrUnlock = lock === "true" ? "lock" : "unlock";
      const infoMsg = `Accepted VM ${lockOrUnlock}`;
      this.$emit("push-toast", { message: infoMsg, color: "is-info" });

      const url = this.endpoint + `vms/${this.item.name}`;
      const body = { lock: lock };
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
            this.$emit("delete-vm", this.item);
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
