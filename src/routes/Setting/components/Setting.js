import React from 'react';
import Button from 'antd/es/button';

function Setting(props) {
  return (
    <div className="App">
      <header className="App-header">
        <p>
          Setting {props.counter}
          <Button onClick={props.increment}>click</Button>
          <Button onClick={props.incrementAsync}>double click</Button>
        </p>
      </header>
    </div>
  );
}

export default Setting;
