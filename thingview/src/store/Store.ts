// Vue based in memory reactive storage
// See: https://medium.com/@mario.brendel1990/vue-3-the-new-store-a7569d4a546f
import { reactive, readonly } from 'vue';


export default abstract class Store<T extends Object> {
  protected state: T;

  constructor() {
    let data = this.data();
    this.setup(data);
    // make the store data reactive
    this.state = reactive(data) as T;
  }

  protected abstract data(): T

  protected setup(data: T): void { }

  public getState(): T {
    // only allow mutation within the provided class
    return readonly(this.state) as T
  }
}

