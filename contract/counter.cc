#include "xchain/xchain.h"
#include <string>

struct Counter : public xchain::Contract {};

DEFINE_METHOD(Counter, initialize) {
    xchain::Context* ctx = self.context();
    const std::string& creator = ctx->arg("creator");
    if (creator.empty()) {
        ctx->error("missing creator");
        return;
    }
    ctx->put_object("creator", creator);
    ctx->ok("initialize succeed");
}

DEFINE_METHOD(Counter, increase) {
    xchain::Context* ctx = self.context();
    const std::string& key = ctx->arg("key");
    std::string value;
    ctx->get_object(key, &value);
    int cnt = 0;
    cnt = atoi(value.c_str());
    ctx->logf("get value %s -> %d", key.c_str(), cnt);
    char buf[32];
    snprintf(buf, 32, "%d", cnt + 1);
    ctx->put_object(key, buf);

    ctx->emit_event("increase", buf);

    ctx->ok(buf);
}

DEFINE_METHOD(Counter, get) {
    xchain::Context* ctx = self.context();
    const std::string& key = ctx->arg("key");
    std::string value;
    if (ctx->get_object(key, &value)) {
        ctx->ok(value);
    } else {
        ctx->error("key not found");
    }
}

DEFINE_METHOD(Counter, setPreKeys) {
    xchain::Context* ctx = self.context();
    const std::string& value = ctx->arg("value");
    for (int i = 0; i < 10; i++)
    {
        std::string s = std::to_string(i);
        ctx->put_object(s, value);
    }

    for (int i = 11; i < 20; i++)
    {
        std::string value;
        std::string s = std::to_string(i);
        ctx->get_object(s, &value);
    }
}

DEFINE_METHOD(Counter, setSuffKeys) {
    xchain::Context* ctx = self.context();
    const std::string& value = ctx->arg("value");
    for (int i = 10; i < 20; i++)
    {
        std::string s = std::to_string(i);
        ctx->put_object(s, value);
    }

    for (int i = 0; i < 10; i++)
    {
        std::string value;
        std::string s = std::to_string(i);
        ctx->get_object(s, &value);
    }
}

DEFINE_METHOD(Counter, getAllKeys) {
    xchain::Context* ctx = self.context();
    std::string result = "";
    for (int i = 0; i < 20; i++)
    {
        std::string value;
        std::string s = std::to_string(i);
        ctx->get_object(s, &value);
        result = result + value;
    }
    ctx->ok(result);
}

// 下面是 sender 相关

DEFINE_METHOD(Counter, setPreKeysWithSender) {
    xchain::Context* ctx = self.context();
    const std::string& caller = ctx->initiator();
    const std::string& value = ctx->arg("value");
    for (int i = 0; i < 10; i++)
    {
        std::string s = std::to_string(i);
        ctx->put_object(s+caller, value);
    }

    for (int i = 11; i < 20; i++)
    {
        std::string value;
        std::string s = std::to_string(i);
        ctx->get_object(s+caller, &value);
    }
}

DEFINE_METHOD(Counter, setSuffKeysWithSender) {
    xchain::Context* ctx = self.context();
    const std::string& caller = ctx->initiator();
    const std::string& value = ctx->arg("value");
    for (int i = 10; i < 20; i++)
    {
        std::string s = std::to_string(i);
        ctx->put_object(s+caller, value);
    }

    for (int i = 0; i < 10; i++)
    {
        std::string value;
        std::string s = std::to_string(i);
        ctx->get_object(s+caller, &value);
    }
}

DEFINE_METHOD(Counter, getAllKeysWithSender) {
    xchain::Context* ctx = self.context();
    const std::string& caller = ctx->initiator();
    std::string result = "";
    for (int i = 0; i < 20; i++)
    {
        std::string value;
        std::string s = std::to_string(i);
        ctx->get_object(s+caller, &value);
        result = result + value;
    }
    ctx->ok(result);
}