import { mount } from 'svelte'
import App from './App.svelte'
import DriveGuide from './DriveGuide.svelte'

const page = new URLSearchParams(window.location.search).get('page')
const target = document.getElementById('app')!

if (page === 'drive-guide') {
  mount(DriveGuide, { target })
} else {
  mount(App, { target })
}
