#include <stdint.h>

namespace {
  volatile uint8_t& memory(const uint16_t loc) {
    return *reinterpret_cast<uint8_t*>(loc);
  }

  constexpr bool test_bit(const uint8_t data, const uint8_t bit) {
    return (data & (1 << bit)) != 0;
  }

/*
  uint16_t row_address(uint8_t row) {
    return 16384 + ((row / 64) * 2048) + ((row % 8) * 256) + (((row / 8) % 8) * 32);
  }

  uint16_t pixel_address(uint8_t row, uint8_t col) {
    return row_address(row) + col;
  }
  */

  struct Joystick {
    // TODO: figure out where in(31) is in memory.
    static constexpr uint16_t JOYSTICK_ADDRESS = 0;

    struct PortData {
      uint8_t data;
    };

    Joystick()
      : Joystick(PortData{memory(JOYSTICK_ADDRESS)}) {}

     Joystick(const PortData d)
       : right(test_bit(d.data, 0)),
         left(test_bit(d.data, 1)),
         down(test_bit(d.data, 2)),
         up(test_bit(d.data, 3)),
         fire(test_bit(d.data, 4)) {}

      bool right, left, down, up, fire;
  };

  struct Z80 {
    enum Color: uint8_t {
      BLACK = 0,
      BLUE = 1,
      RED = 2,
      MAGENTA = 3,
      GREEN = 4,
      CYAN = 5,
      YELLOW = 6,
      WHITE = 7,
    };

    using PixelColor = uint8_t;

    // static constexpr uint16_t SCREEN_BASE = 16384;
    static constexpr uint16_t SCREEN_COLOR_BASE = 22528;

    // TODO: figure out where out (254) is in memory.
    static constexpr uint16_t BORDER_ADDRESS = 100;

    static PixelColor color(const Color ink, const Color paper,
                            const bool bright, const bool flash) {
      return PixelColor(
          uint8_t(ink) + (uint8_t(paper) * 8) + ((bright?1:0) * 64) + ((flash?1:0) * 128));
    }

    void set_color(const uint8_t x, const uint8_t y, const PixelColor c) {
      memory(SCREEN_COLOR_BASE + x + (32*y)) = c;
    }

    void border(const Color c) {
      memory(BORDER_ADDRESS) = c;
    }
  };

  struct Player {
    Player(Z80* z80, const uint8_t x, const uint8_t y, const Z80::PixelColor c)
      : z80(z80), x(x), y(y), c(c) {}

    void Draw() {
      z80->set_color(x, y, c);
    }

    Z80* const z80;

    uint8_t x, y;
    const Z80::PixelColor c;
  };
}

int main() {
  Z80 z80;
  Player p1(&z80, 2, 12, Z80::color(Z80::Color::BLACK, Z80::Color::GREEN,
                                    false, false));

  Player p2(&z80, 30, 12, Z80::color(Z80::Color::BLACK, Z80::Color::RED,
                                     false, false));

  while (true) {
    // if (const auto joy = Joystick(); joy.up) {
    //   if (p1.y < 22) p1.y += 2;
    // } else if(joy.down) {
    //   if (p1.y > 2) p1.y -= 2;
    // }

    p1.y = (p1.y + 1) % 24;
    p1.Draw();
    p2.Draw();
  }
}
