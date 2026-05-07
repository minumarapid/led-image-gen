/* @refresh reload */
import { render } from 'solid-js/web'
import './index.css'
import App from './App.tsx'
import '@fontsource-variable/plus-jakarta-sans/wght.css';
import '@fontsource-variable/m-plus-2/wght.css';

const root = document.getElementById('root')

render(() => <App />, root!)
