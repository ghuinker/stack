import mitt from 'mitt'
import { onMounted, onUnmounted } from 'vue'

const emitter = mitt()

const oldOn = emitter.on

emitter.on = (name, fun) => {
  onMounted(() => oldOn(name, fun))
  onUnmounted(() => emitter.off(name, fun))
}

export default emitter
