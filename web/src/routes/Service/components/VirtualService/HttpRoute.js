import React, { Component } from "react";
import { Tag, Row, Col, Table } from 'antd';

export const MatchConig = (props) => {
  console.log(props, 'props');
  const { config } = props;
  if (config.uri) {
    return (<Tag>{config.uri.prefix}</Tag>)
  }
  console.log(config);
  return <Tag>Unknown Match Config</Tag>
}

export default class HttpRoute extends Component {
  render() {
    const { match, route } = this.props.httpRoute;
    // console.log(match, route);
    return (
      <Row className="virtualservice-spec-http-route">
        <Col className="match" span={4}>
          {!match ? <span>Default</span> :
            match.map((config, index) => {
              return <MatchConig key={`match-config-${index}`} config={config} />
            })
          }
        </Col>
        <Col className="route" span={12}>
          {route.map((config, index) => {
            const destination = config.destination;
            return (<Tag key={`route-config-${index}`}>{destination.subset}</Tag>)
          })}
        </Col>
      </Row>
    );
  }
}