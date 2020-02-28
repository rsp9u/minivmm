<template lang="pug">
  #minivmm_fwdialog
    form
      .modal-card(style="max-width: 32em")
        header.modal-card-head
          p.modal-card-title Add Forwarding rule
        section.modal-card-body
          b-tabs(v-model="activeTab")
            b-tab-item(label="Simple")
              .columns.is-multiline
                .column.is-12
                  b-field(label="host")
                    b-select(v-model="simple.host" expanded)
                      option(v-for="option in allHosts" :key="option" :value="option") {{ option }}
                .column.is-12
                  b-field(label="port(tcp)")
                    b-input(v-model="simple.to_port")
                .column.is-12
                  b-field(label="type")
                    b-select(v-model="simple.type" expanded)
                      option(v-for="option in types" :key="option" :value="option") {{ option }}
            b-tab-item(label="Expert")
              .columns.is-multiline
                .column.is-12
                  b-field(label="hypervisor")
                    b-select(v-model="expert.hypervisor" expanded)
                      option(v-for="option in agentNames" :key="option" :value="option") {{ option }}
                .column.is-3
                  b-field(label="protocol")
                    b-select(v-model="expert.proto" expanded)
                      option(v-for="option in protocols" :key="option" :value="option") {{ option }}
                .column.is-3
                  b-field(label="from port")
                    b-input(v-model="expert.from_port")
                .column.is-3
                  b-field(label="to name")
                    b-select(v-model="expert.to_name" expanded)
                      option(v-for="option in vmNames" :key="option" :value="option") {{ option }}
                .column.is-3
                  b-field(label="to port")
                    b-input(v-model="expert.to_port")
                .column.is-6
                  b-field(label="description")
                    b-input(v-model="expert.description")
                .column.is-6
                  b-field(label="type")
                    b-select(v-model="expert.type" expanded)
                      option(v-for="option in types" :key="option" :value="option") {{ option }}
        footer.modal-card-foot(style="justify-content: flex-end")
          b-button(@click="cancel") Cancel
          b-button(type="is-info" @click="create") Create
</template>

<script>
export default {
  name: "FWDialog",
  props: ["agents", "vms"],
  data() {
    return {
      activeTab: 0,
      protocols: ["tcp", "udp"],
      types: ["http/https", "ssh"],
      simple: { proto: "tcp" },
      expert: { proto: "tcp" }
    };
  },
  computed: {
    agentNames() {
      return this.agents.map(x => x.name);
    },
    vmNames() {
      return this.vms.filter(x => x.hypervisor === this.expert.hypervisor).map(x => x.name);
    },
    allHosts() {
      return this.vms.map(x => `${x.hypervisor}:${x.name}`);
    }
  },
  methods: {
    clear() {
      this.simple = { proto: "tcp" };
      this.expert = { proto: "tcp" };
    },
    cancel() {
      this.$emit("cancel");
      this.clear();
    },
    create() {
      var fw;
      if (this.activeTab === 0) {
        fw = {
          hypervisor: this.simple.host.split(":")[0],
          proto: "tcp",
          to_name: this.simple.host.split(":")[1],
          to_port: this.simple.to_port,
          description: `${this.simple.type} to ${this.simple.host}`,
          type: this.simple.type
        };
      } else {
        fw = this.expert;
      }
      this.$emit("create", fw);
    }
  }
}
</script>
