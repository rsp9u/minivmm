<template lang="pug">
  #minivmm_resource_graph.section
    template(v-for="r in displayResources")
      .columns
        .column.is-1
          b {{ r.name }}:
        .column
          template(v-if="!r.error")
            .columns
              .column
                b-progress(size="is-medium" show-value :value="r.cpu.value"  :max="r.cpu.max"  :type="r.cpu.color")
                  span(style="color: black") {{ r.cpu.text }}
              .column
                b-progress(size="is-medium" show-value :value="r.mem.value"  :max="r.mem.max"  :type="r.mem.color")
                  span(style="color: black") {{ r.mem.text }}
              .column
                b-progress(size="is-medium" show-value :value="r.disk.value" :max="r.disk.max" :type="r.disk.color")
                  span(style="color: black") {{ r.disk.text }}
          template(v-else)
            b(style="color: red") Its resource information is unavailable.
</template>

<script>
export default {
  name: "ResourceGraph",
  props: ["resources"],
  computed: {
    displayResources() {
      return this.resources.map((r) => {
        try {
          const memUsedMB = r.res.sys.minivmm_sys_memory_bytes_used / 1024 / 1024;
          const memTotalMB = r.res.sys.minivmm_sys_memory_bytes / 1024 / 1024;
          const diskUsedGB = r.res.sys.minivmm_sys_disk_bytes_used / 1024 / 1024 / 1024;
          const diskTotalGB = r.res.sys.minivmm_sys_disk_bytes / 1024 / 1024 / 1024;

          const cpuColor = r.res.vm.minivmm_cpu_cores_running > r.res.sys.minivmm_sys_cpu_cores ? "is-danger" : "is-success";
          const memColor = memUsedMB / memTotalMB > 0.9 ? "is-danger" : memUsedMB / memTotalMB > 0.75 ? "is-warning" : "is-success";
          const diskColor = diskUsedGB / diskTotalGB > 0.9 ? "is-danger" : diskUsedGB / diskTotalGB > 0.75 ? "is-warning" : "is-success";

          return {
            name: r.name,
            cpu: {
              value: r.res.vm.minivmm_cpu_cores_running,
              max: r.res.sys.minivmm_sys_cpu_cores,
              text: `cpu: ${r.res.vm.minivmm_cpu_cores_running}/${r.res.sys.minivmm_sys_cpu_cores}`,
              color: cpuColor,
            },
            mem: {
              value: r.res.sys.minivmm_sys_memory_bytes_used,
              max: r.res.sys.minivmm_sys_memory_bytes,
              text: `memory: ${memUsedMB.toFixed(0)}/${memTotalMB.toFixed(0)}[MB]`,
              color: memColor,
            },
            disk: {
              value: r.res.sys.minivmm_sys_disk_bytes_used,
              max: r.res.sys.minivmm_sys_disk_bytes,
              text: `disk: ${diskUsedGB.toFixed(0)}/${diskTotalGB.toFixed(0)}[GB]`,
              color: diskColor,
            },
            error: false,
          }
        } catch {
          return {name: r.name, error: true};
        }
      });
    }
  }
}
</script>

<style scoped>
.section {
  padding-bottom: 1.0rem;
}
</style>
