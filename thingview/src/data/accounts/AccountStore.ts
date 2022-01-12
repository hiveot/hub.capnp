// HubAccountStore is a local storage persistent data for hub accounts
import { reactive, readonly } from "vue";
import { nanoid } from 'nanoid'

// Hub Account record
export class AccountRecord extends Object {
  // unique account id (required)
  id: string = nanoid(5);

  // Account friendly name for display
  name: string = "new account";

  // login credentials
  loginName: string = "email@something";

  // Hub hostname or IP address (must match its server certificate name)
  address: string = "localhost";

  // port of authentication service
  authPort?: number = 8881;

  // port of mqtt service. 8884 for certificate auth, 8885 for websocket 
  mqttPort?: number = 8885;

  // port of the directory service
  directoryPort?: number = 8886;

  // when enabled, attempt to connect
  enabled: boolean = false;
}

class AccountsData {
  accounts = Array<AccountRecord>()

  // RememberMe persists account(s) in local storage on Save.
  // If enabled, the refresh token will be saved in a secure cookie.
  rememberMe: boolean = false
}

// Hub account data implementation with additional methods for loading and saving
export class AccountStore {
  data: AccountsData
  storageKey: string = "accountStore"

  constructor() {
    let defaultAccount = new AccountRecord()
    defaultAccount.name = "Hub server"
    defaultAccount.address = location.hostname
    defaultAccount.loginName = "user1" // for testing
    defaultAccount.enabled = true
    this.data = reactive(new AccountsData())
  }

  // add a new account to the list and save the account list
  Add(account: AccountRecord): void {
    // ensure each account has an ID
    if (!account.id) {
      account.id = nanoid(5)
    }
    let newAccount = JSON.parse(JSON.stringify(account))
    this.data.accounts.push(newAccount)
    this.Save()
  }

  // Return a list of accounts
  get accounts(): readonly AccountRecord[] {
    return readonly(this.data.accounts) as AccountRecord[]
  }

  // Get the current status of rememberMe
  get rememberMe(): boolean {
    return this.data.rememberMe
  }
  // Set the current status of rememberMe
  set rememberMe(remember) {
    this.data.rememberMe = remember
  }

  // Get the account with the given id
  GetAccountById(id: string): AccountRecord | undefined {
    let accounts = this.data.accounts

    let ac = accounts.find(el => (el.id === id))
    if (!ac) {
      return undefined
    }
    return readonly(ac) as AccountRecord

  }

  /** Load accounts from session/local storage
   * First load from session storage. If session storage is empty, try local storage
   */
  Load() {
    let serializedStore = sessionStorage.getItem(this.storageKey)
    if (!serializedStore) {
      serializedStore = localStorage.getItem(this.storageKey)
    }
    if (serializedStore != null) {
      let accountData: AccountsData = JSON.parse(serializedStore)
      if (accountData != null && accountData.accounts.length > 0) {
        this.data.accounts.splice(0, this.data.accounts.length)
        this.data.accounts.push(...accountData.accounts)
        console.debug("Loaded %s accounts from local storage", accountData.accounts.length)
      } else {
        console.log("No accounts in storage. Keeping existing accounts")
      }
    }
    // ensure there is at least 1 account to display
    if (this.data.accounts.length == 0) {
      let defaultAccount = new AccountRecord()
      this.data.accounts.push(defaultAccount)
    }
  }

  // remove the given account by id
  Remove(id: string) {
    let remainingAccounts = this.data.accounts.filter((item: AccountRecord) => {
      // console.log("Compare id '",id,"' with item id: ", item.id)
      return (item.id != id)
    })
    console.log("Removing account with id", id,)
    this.data.accounts.splice(0, this.data.accounts.length)
    this.data.accounts.push(...remainingAccounts)
    this.Save()
  }

  // Save account in session storage
  // If RememberMe is set, also save the data in localStorage for use between sessions.
  // If RememberMe is not set, the localstorage key is removed
  Save() {
    console.log("Saving %s accounts to local storage", this.data.accounts.length)
    let serializedStore = JSON.stringify(this.data)
    sessionStorage.setItem(this.storageKey, serializedStore)
    if (this.data.rememberMe) {
      localStorage.setItem(this.storageKey, serializedStore)
    } else {
      localStorage.removeItem(this.storageKey)
    }
  }

  // Enable or disable the hub account
  // When enabled is true, an attempt will be made to connect to the Hub on the port(s)
  // When enabled is false, any existing connections will be closed
  SetEnabled(id: string, enabled: boolean) {
    let account = this.data.accounts.find(el => (el.id === id))
    if (!account) {
      console.log("SetEnabled: ERROR account with ID", id, " not found")
      return
    }
    console.log("SetEnabled of account", account.name, ":", enabled)
    account.enabled = enabled
    this.Save()

  }

  // Update the account with the given record and save
  // If the record ID does not exist, and ID will be assigned and the record is added
  // If the record ID exists, the record is updated
  Update(account: AccountRecord) {
    let newAccount = JSON.parse(JSON.stringify(account))

    let existing = this.data.accounts.find(el => (el.id === account.id))

    if (!existing) {
      console.log("Adding account", newAccount)
      this.data.accounts.push(newAccount) // why would it not exist?
    } else {
      console.log("Update account", newAccount)
      // reactive update of the existing record
      Object.assign(existing, newAccount)
    }
    this.Save()
  }
}

// accountStore is a singleton
let accountStore = new AccountStore()

export default accountStore
