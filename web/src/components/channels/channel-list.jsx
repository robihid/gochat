import React, { Component } from 'react';
import PropTypes from 'prop-types';

import Channel from './channel';

class ChannelList extends Component {
  render() {
    return (
      <ul>
        {this.props.channels.map(channel => (
          <Channel channel={channel} key={channel.id} {...this.props} />
        ))}
      </ul>
    );
  }
}

// PropTypes
ChannelList.propTypes = {
  channels: PropTypes.array.isRequired,
  setChannel: PropTypes.func.isRequired,
  activeChannel: PropTypes.object.isRequired
};

export default ChannelList;
