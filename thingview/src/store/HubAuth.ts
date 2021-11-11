// HubAuth is a non-persistent store for authentication token that supports token refresh
import Store from './Store'


// Hub Authentication data to be stored 
class HubAuthData extends Object {

  // Current authentication status
  isAuthenticated: boolean = false;

  loginName: string = '';
  accessToken: string = '';
  refreshToken: string = '';


}


// HubAuth implements the authentication store
class HubAuth extends Store<HubAuthData> {

  protected data(): HubAuthData {
    return {
      isAuthenticated: false,
      loginName: "",
      accessToken: "",
      refreshToken: "",
    };
  }

  // login to the Hub.
  // This obtains an access and refresh token
  //
  // username and password are provided by the administrator
  // rememberMe to store the refresh token in a secure cookie for persistence between sessions
  // returns a promise with the authentication result
  login(username: String, password: string, rememberMe: boolean): Promise<boolean> {
    return new Promise((resolve, reject) => {
      this.state.isAuthenticated = true
      resolve(this.getState().isAuthenticated);
    })
  };

  // logout from the Hub
  // This clears the access and refresh tokens and the secure cookie
  logout() {
  }

  // refresh the access token
  refresh(): Promise<void> {
    return new Promise((resolve, reject) => {
      resolve()
    });
  }

}

export const hubAuth: HubAuth = new HubAuth()