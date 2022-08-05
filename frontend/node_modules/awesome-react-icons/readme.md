<h1 align="center">awesome-react-icons</h1>
<p align="center">  </p>
<div align="center">
  <a href="https://www.npmjs.com/package/awesome-react-icons">
    <img src="https://img.shields.io/npm/v/awesome-react-icons" alt="NPM" />
  </a>
  <a href="https://travis-ci.com/abhijithvijayan/awesome-react-icons">
    <img src="https://travis-ci.com/abhijithvijayan/awesome-react-icons.svg?branch=main" alt="Travis Build" />
  </a>
  </a>
  <a href="https://david-dm.org/abhijithvijayan/awesome-react-icons">
    <img src="https://img.shields.io/david/abhijithvijayan/awesome-react-icons.svg?colorB=orange" alt="DEPENDENCIES" />
  </a>
  <a href="https://github.com/abhijithvijayan/awesome-react-icons/blob/main/license">
    <img src="https://img.shields.io/github/license/abhijithvijayan/awesome-react-icons.svg" alt="LICENSE" />
  </a>
  <a href="https://twitter.com/intent/tweet?text=Check%20out%20awesome-react-icons%21%20by%20%40_abhijithv%0A%0AYet%20another%20minimal%20icons%20library%0Ahttps%3A%2F%2Fgithub.com%2Fabhijithvijayan%2Fawesome-react-icons%0A%0A%23react%20%23icons%20%23library%20%23javascript%20%23typescript">
     <img src="https://img.shields.io/twitter/url/http/shields.io.svg?style=social" alt="TWEET" />
  </a>
</div>
<h3 align="center">🙋‍♂️ Made by <a href="https://twitter.com/_abhijithv">@abhijithvijayan</a></h3>
<p align="center">
  Donate:
  <a href="https://www.paypal.me/iamabhijithvijayan" target='_blank'><i><b>PayPal</b></i></a>,
  <a href="https://www.patreon.com/abhijithvijayan" target='_blank'><i><b>Patreon</b></i></a>
</p>
<p align="center">
  <a href='https://www.buymeacoffee.com/abhijithvijayan' target='_blank'>
    <img height='36' style='border:0px;height:36px;' src='https://bmc-cdn.nyc3.digitaloceanspaces.com/BMC-button-images/custom_images/orange_img.png' border='0' alt='Buy Me a Coffee' />
  </a>
</p>
<hr />

❤️ it? ⭐️ it on [GitHub](https://github.com/abhijithvijayan/awesome-react-icons/stargazers) or [Tweet](https://twitter.com/intent/tweet?text=Check%20out%20awesome-react-icons%21%20by%20%40_abhijithv%0A%0AYet%20another%20minimal%20icons%20library%0Ahttps%3A%2F%2Fgithub.com%2Fabhijithvijayan%2Fawesome-react-icons%0A%0A%23react%20%23icons%20%23library%20%23javascript%20%23typescript) about it.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [API](#api)
- [Issues](#issues)
  - [🐛 Bugs](#-bugs)
- [LICENSE](#license)

## Installation

Ensure you have [Node.js](https://nodejs.org) 10 or later installed. Then run the following:

```
# via npm
npm install awesome-react-icons

# or yarn
yarn add awesome-react-icons
```

## Usage

```js
import Icon from 'awesome-react-icons';

const SampleIcon = () => {
  return (
    <Icon name="user" />
  )
}
```

## API

#### name

Type: `string`

Icon's name

#### size

Type: `number` or `string`

Default: `20`

Size of icon

#### stroke

Type: `string`

Default: `currentColor`

Stroke color

#### strokeWidth

Type: `number` or `string`

Default: `1.5`

Stroke width

#### rootDivClassName

Type: `string`

Default: `icon-wrapper`

Wrapper div element's class name

## Issues

_Looking to contribute? Look for the [Good First Issue](https://github.com/abhijithvijayan/awesome-react-icons/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3A%22good+first+issue%22)
label._

### 🐛 Bugs

Please file an issue [here](https://github.com/abhijithvijayan/awesome-react-icons/issues/new) for bugs, missing documentation, or unexpected behavior.

[**See Bugs**](https://github.com/abhijithvijayan/awesome-react-icons/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3A%22type%3A+bug%22)

### Linting & TypeScript Config

- Shared Eslint & Prettier Configuration - [`@abhijithvijayan/eslint-config`](https://www.npmjs.com/package/@abhijithvijayan/eslint-config)
- Shared TypeScript Configuration - [`@abhijithvijayan/tsconfig`](https://www.npmjs.com/package/@abhijithvijayan/tsconfig)

## License

MIT © [Abhijith Vijayan](https://abhijithvijayan.in)
