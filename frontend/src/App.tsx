import { Route, Switch } from 'react-router-dom'
import { ThemeProvider } from './context/ThemeContext'
import { ToastProvider } from './context/ToastContext'
import { UserProvider } from './context/UserContext'
import { AppLayout } from './layouts/AppLayout'
import { Downloads } from './pages/Downloads'
import { Groups } from './pages/Groups'
import { Home } from './pages/Home'
import { Settings } from './pages/Settings'
import { Tutorial } from './pages/Tutorial'

export default function App() {
  return (
    <ThemeProvider>
      <ToastProvider>
        <UserProvider>
          <AppLayout>
            <Switch>
              <Route exact path="/">
                <Home />
              </Route>
            <Route exact path="/downloads">
              <Downloads />
            </Route>
            <Route exact path="/groups">
              <Groups />
            </Route>
            <Route exact path="/tutorial">
                <Tutorial />
              </Route>
              <Route exact path="/settings">
                <Settings />
              </Route>
            </Switch>
          </AppLayout>
        </UserProvider>
      </ToastProvider>
    </ThemeProvider>
  )
}
