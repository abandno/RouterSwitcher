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
window.addEventListener('beforeunload', async function (e) {
    console.log('beforeunload')
    // 使用 runtime API 直接调用后端 HideWindow 方法
    // 这样可以避免导入绑定文件的依赖问题
    try {
        // if (window.runtime && window.runtime.go && window.runtime.go.main && window.runtime.go.main.WailsApp) {
        //     const hideWindowPromise = window.runtime.go.main.WailsApp.HideWindow();
        //     if (hideWindowPromise && typeof hideWindowPromise.catch === 'function') {
        //         hideWindowPromise.catch(err => {
        //             console.error('隐藏窗口失败:', err);
        //         });
        //     }
        // } else {
        //     console.warn('无法访问 runtime API');
        // }
        try {
            await HideWindow()
            alert('隐藏窗口成功')
        } catch (err) {
            console.error('隐藏窗口失败:', err)
            alert('隐藏窗口失败: ' + err)
        }
    } catch (err) {
        console.error('调用 HideWindow 失败:', err);
    }
    // 阻止默认的关闭行为
    e.preventDefault();
    e.returnValue = ''; // Chrome 需要这个
    return ''; // 某些浏览器需要返回值
});