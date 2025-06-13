import { useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'
import { appName, appVersion, environment, apiUrl, features, isFeatureEnabled } from './config/config'

function App() {
  const [count, setCount] = useState(0)

  return (
    <>
      <div>
        <a href="https://vitejs.dev" target="_blank">
          <img src={viteLogo} className="logo" alt="Vite logo" />
        </a>
        <a href="https://react.dev" target="_blank">
          <img src={reactLogo} className="logo react" alt="React logo" />
        </a>
      </div>
      <h1>{appName}</h1>
      <div className="card">
        <button onClick={() => setCount((count) => count + 1)}>
          count is {count}
        </button>
        <p>
          Edit <code>src/App.tsx</code> and save to test HMR
        </p>
      </div>
      
      {/* Configuration Information */}
      <div className="config-info" style={{ marginTop: '2rem', padding: '1rem', border: '1px solid #ccc', borderRadius: '8px' }}>
        <h2>Configuration Information</h2>
        <div style={{ textAlign: 'left', fontSize: '0.9rem' }}>
          <p><strong>App:</strong> {appName} v{appVersion}</p>
          <p><strong>Environment:</strong> {environment}</p>
          <p><strong>API URL:</strong> {apiUrl}</p>
          <p><strong>Features Enabled:</strong></p>
          <ul style={{ margin: '0.5rem 0', paddingLeft: '2rem' }}>
            {Object.entries(features).map(([feature, enabled]) => (
              <li key={feature} style={{ color: enabled ? 'green' : 'red' }}>
                {feature}: {enabled ? '✅' : '❌'}
              </li>
            ))}
          </ul>
          {isFeatureEnabled('debugMode') && (
            <div style={{ marginTop: '1rem', padding: '0.5rem', backgroundColor: '#f0f0f0', borderRadius: '4px' }}>
              <strong>Debug Mode Active</strong>
              <p>Full config object logged to console</p>
            </div>
          )}
        </div>
      </div>
      
      <p className="read-the-docs">
        Click on the Vite and React logos to learn more
      </p>
    </>
  )
}

export default App
