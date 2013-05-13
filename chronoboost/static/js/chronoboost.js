function LifeF1() {
    this._init();
}

LifeF1.prototype = {
    _init : function() {
        this._carTable = new CarTable(document.querySelector("#cartable"));
        this._weatherView = new WeatherView(document.querySelector("#weather"));
        this._ws = new WebSocket("ws://127.0.0.1:8080/ws");
        this._ws.onmessage = this._onMessage.bind(this);
    },

    _onMessage : function(messageEvent) {
        console.log("Received message " + messageEvent.data);
        var message = JSON.parse(messageEvent.data);

        if (!('Type' in message) || !('Value' in message)) {
            console.log("Received invalid message " + messageEvent.data);
            return;
        }

        switch (message.Type) {
            case "car":
                this._carTable.updateFromCar(message.Value);
                break;
            case "weather":
                this._weatherView.updateFromWeather(message.Value);
                break;
            default:
                break;
        }

    }
};

var liveF1 = new LifeF1();
