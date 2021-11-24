// HubAccountStore is a local storage persistent store for hub accounts
import Store from './Store'
import {reactive} from "vue";


// Single account record
class AccountRecord extends Object {

  // Account name
  name: string = "";
  // Hub hostname or IP address
  address: string = "localhost";
  mqttPort: number = 8883;
  directoryPort: number = 9678;
  enabled: boolean = false;
}

// API for account storage client
interface IAccountStore {
  // Add a new account
  // This replaces an existing account with the same name
  Add(account: AccountRecord): void;

  // remove the account with the given name
  Remove(name: string): void;

  // Return a reactive array of accounts
  GetAccounts(): AccountRecord[];

  // Enable/disable account
  SetEnabled(name: string, enabled: boolean): void
}

// Hub account store implementation with additional methods for loading and saving
class AccountStore<IAccountStore> {
  state: {
    accounts: Array<AccountRecord>
  }
  storageKey: string = "hubAccountStore"

  constructor() {
    this.state = reactive( {
      accounts: [new AccountRecord()]
    })
  }

  // add a new account to the list
  Add(account: AccountRecord):void {
    this.state.accounts.push(account)
  }

  // Return a list of accounts
  GetAccounts(): AccountRecord[] {
    return this.state.accounts
  }

  // Get the account with the given name or null if not found
  GetAccountByName(name: string): AccountRecord|undefined {
    let accounts = this.state.accounts

    let el = accounts.find( el => el.name == name)
    return el
  }

  // load accounts from local storage
  Load() {
    let serializedStore = localStorage.getItem(this.storageKey)
    if (serializedStore != null) {
      let accountList:AccountRecord[] = JSON.parse(serializedStore)
      if (accountList != null) {
        this.state.accounts.splice(0, this.state.accounts.length)
        this.state.accounts.push(...accountList )
        console.debug("Loaded %s accounts from local storage", accountList.length)
      } else {
        console.log("No accounts in storage")
      }
    }
  }
  // remove the given account name
  Remove(name: string) {
    let remainingAccounts = this.state.accounts.filter((item:AccountRecord) => {
      return (item.name != name)
    })
    this.state.accounts.splice(0, this.state.accounts.length)
    this.state.accounts.push(...remainingAccounts )
  }

  // save to local storage
  Save() {
    console.log("Saving %s accounts to local storage", this.state.accounts.length)
    let serializedStore = JSON.stringify(this.state)
    localStorage.setItem(this.storageKey, serializedStore)
  }

  // Enable or disable the hub account
  // When enabled is true, an attempt will be made to connect to the Hub on the port(s)
  // When enabled is false, any existing connections will be closed
  SetEnabled(name: string, enabled:boolean) {
    let account = this.GetAccountByName(name)
    if (account) {
      console.log("SetEnabled of account", name, ":", enabled)
      account.enabled = enabled
    } else {
      console.log("SetEnabled of account", name, ": ERROR not found")
    }
  }
}

// hubAccountStore is a singleton
let hubAccountStore = new AccountStore()

export { IAccountStore, AccountRecord };
export default hubAccountStore
