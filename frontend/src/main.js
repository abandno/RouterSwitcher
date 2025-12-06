import './style.css';
import './app.css';

import { createApp } from 'vue';
import ConfigManager from './components/ConfigManager.vue';
import logo from './assets/images/logo-universal.png';
import {Greet, HideWindow} from '../bindings/RouterSwitcher/wailsapp';

document.querySelector('#app').innerHTML = `
    <img id="logo" class="logo">
      <div class="result" id="result">Please enter your name below 👇</div>
      <div class="input-box" id="input">
        <input class="input" id="name" type="text" autocomplete="off" />
        <button class="btn" onclick="greet()">Greet</button>
      </div>
    </div>
`;
document.getElementById('logo').src = logo;

let nameElement = document.getElementById("name");
nameElement.focus();
let resultElement = document.getElementById("result");

// Setup the greet function
window.greet = function () {
    // Get name
    let name = nameElement.value;

    // Check if the input is empty
    if (name === "") return;

    // Call App.Greet(name)
    try {
        Greet(name)
            .then((result) => {
                // Update result with data back from App.Greet()
                resultElement.innerText = result;
            })
            .catch((err) => {
                console.error(err);
            });
    } catch (err) {
        console.error(err);
    }
};

const app = createApp(ConfigManager);
app.mount('#app');

// 拦截窗口关闭事件，隐藏窗口而不是关闭
// 在 Wails v3 中，通过 beforeunload 事件拦截窗口关闭
window.addEventListener('beforeunload', async function (e) {
    console.log('beforeunload 事件触发 - 拦截窗口关闭，改为隐藏')
    // 阻止默认的关闭行为
    e.preventDefault();
    e.returnValue = ''; // Chrome 需要这个
    
    // 调用 Go 端的 HideWindow 方法隐藏窗口
    try {
        await HideWindow()
        console.log('窗口已隐藏')
    } catch (err) {
        console.error('调用 HideWindow 失败:', err);
    }
    
    return ''; // 某些浏览器需要返回值
});