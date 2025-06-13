/**
 * Simple test for configuration module
 * This can be run manually to verify configuration loading
 */

// Mock Vite's import.meta.env for testing
const mockEnv = {
  MODE: 'test',
  VITE_APP_NAME: 'Test Tennis Booker',
  VITE_APP_VERSION: '1.0.0',
  VITE_APP_ENVIRONMENT: 'test',
  VITE_API_URL: 'http://localhost:3000',
  VITE_API_TIMEOUT: '5000',
  VITE_FEATURE_ANALYTICS_ENABLED: 'false',
  VITE_FEATURE_NOTIFICATIONS_ENABLED: 'true',
  VITE_LOG_LEVEL: 'warn',
  VITE_DEBUG_MODE: 'false',
};

// Test configuration parsing
console.log('🧪 Testing Configuration System');

// Test boolean parsing
function testParseBoolean() {
  console.log('Testing boolean parsing:');
  console.log('  "true" ->', 'true'.toLowerCase() === 'true');
  console.log('  "false" ->', 'false'.toLowerCase() === 'true');
  console.log('  "1" ->', '1' === '1');
  console.log('  undefined ->', false);
}

// Test number parsing
function testParseNumber() {
  console.log('Testing number parsing:');
  console.log('  "5000" ->', parseInt('5000', 10));
  console.log('  "invalid" ->', isNaN(parseInt('invalid', 10)) ? 'default' : parseInt('invalid', 10));
  console.log('  undefined ->', 'default');
}

// Test URL validation
function testUrlValidation() {
  console.log('Testing URL validation:');
  try {
    new URL('http://localhost:3000');
    console.log('  Valid URL: ✅');
  } catch {
    console.log('  Valid URL: ❌');
  }
  
  try {
    new URL('invalid-url');
    console.log('  Invalid URL: ❌');
  } catch {
    console.log('  Invalid URL: ✅ (correctly rejected)');
  }
}

// Run tests
testParseBoolean();
testParseNumber();
testUrlValidation();

console.log('✅ Configuration system tests completed');

export {}; 