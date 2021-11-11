// HubAccountStore is a local storage persistent store for hub accounts
import Store from './Store'
import {reactive} from "vue";


// Single account record
class AccountRecord extends Object {

  // Current authentication status
  name: string = "";
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
}


// Hub account store implementation with additional methods for loading and saving
export default {
  state: reactive( [new AccountRecord()] ),

  storageKey: "hubAccountStore",

  // add a new account to the list
  Add(account: AccountRecord):void {
    this.state.push(account)
  },

  GetAccounts(): AccountRecord[] {
    return this.state
  },

  // load accounts from local storage
  Load: function() {
    let serializedStore = localStorage.getItem(this.storageKey)
    if (serializedStore != null) {
      let accountList:AccountRecord[] = JSON.parse(serializedStore)
      if (accountList != null) {
        this.state.splice(0, this.state.length)
        this.state.push(...accountList )
        console.debug("Loaded %s accounts from local storage", accountList.length)
      } else {
        console.log("No accounts in storage")
      }
    }
  },
  // remove the given account name
  Remove(name: string):void {
    let remainingAccounts = this.state.filter((item:AccountRecord) => {
      return (item.name != name)
    })
    this.state.splice(0, this.state.length)
    this.state.push(...remainingAccounts )
  },

  // save to local storage
  Save: function() {
    console.log("Saving %s accounts to local storage", this.state.length)
    let serializedStore = JSON.stringify(this.state)
    localStorage.setItem(this.storageKey, serializedStore)
  }
}

export { IAccountStore, AccountRecord };
