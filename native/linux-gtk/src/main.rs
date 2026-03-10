mod truco_core;
mod models;
mod window;

use std::rc::Rc;
use std::cell::RefCell;
use adw::prelude::*;
use gtk::prelude::*;
use gtk::glib;
use truco_core::TrucoCore;

struct AppState {
    pub core: TrucoCore,
    pub window: window::TrucoWindow,
    pub last_snapshot_str: String,
}

fn main() {
    let app = adw::Application::builder()
        .application_id("dev.truco.Native")
        .build();

    app.connect_startup(|_| {
        load_css();
    });

    app.connect_activate(|app| {
        let core = TrucoCore::new();
        
        let window = window::TrucoWindow::new(app);

        // Start offline game
        let _ = core.dispatch(r#"{"kind":"new_offline_game","payload":{"player_names":["Voce","CPU-2"],"cpu_flags":[false,true],"seed_lo":7,"seed_hi":9}}"#);

        let state = Rc::new(RefCell::new(AppState {
            core,
            window: window.clone(),
            last_snapshot_str: String::new(),
        }));

        // Game loop polling
        glib::timeout_add_local(std::time::Duration::from_millis(50), glib::clone!(#[strong] state, move || {
            let mut s = state.borrow_mut();
            if let Some(_event) = s.core.poll_event() {
                // Event pumped, let's refresh snap
            }
            if let Some(snap_str) = s.core.snapshot() {
                if snap_str != s.last_snapshot_str {
                    s.last_snapshot_str = snap_str.clone();
                    if let Ok(snap) = serde_json::from_str::<models::GameSnapshot>(&snap_str) {
                        update_ui(&s.window, &snap, &s.core);
                    }
                }
            }
            glib::ControlFlow::Continue
        }));

        window.present();
    });

    app.run();
}

fn load_css() {
    let provider = gtk::CssProvider::new();
    provider.load_from_data(include_str!("../style.css"));
    gtk::style_context_add_provider_for_display(
        &gtk::gdk::Display::default().expect("Could not connect to a display."),
        &provider,
        gtk::STYLE_PROVIDER_PRIORITY_APPLICATION,
    );
}

fn update_ui(window: &window::TrucoWindow, snapshot: &models::GameSnapshot, core: &TrucoCore) {
    // 3. HUD Layer
    let hud_box = window.hud_box();
    while let Some(child) = hud_box.first_child() {
        hud_box.remove(&child);
    }
    
    if let Some(pts) = &snapshot.match_points {
        let score_lbl = gtk::Label::new(None);
        score_lbl.set_markup(&format!("<span size='x-large'><b>NÓS</b></span>\n<span size='30000'><b>{}</b></span>", pts[0]));
        score_lbl.set_justify(gtk::Justification::Center);
        score_lbl.add_css_class("score-hud");
        
        let stake_lbl = gtk::Label::new(None);
        let stake = snapshot.current_hand.as_ref().and_then(|h| h.stake).unwrap_or(1);
        stake_lbl.set_markup(&format!("<span size='x-large'><b>VALE</b></span>\n<span size='30000' color='#ffcc00'><b>{}</b></span>", stake));
        stake_lbl.set_justify(gtk::Justification::Center);
        stake_lbl.add_css_class("stake-badge");
        
        let score_them = gtk::Label::new(None);
        score_them.set_markup(&format!("<span size='x-large'><b>ELES</b></span>\n<span size='30000'><b>{}</b></span>", pts[1]));
        score_them.set_justify(gtk::Justification::Center);
        score_them.add_css_class("score-hud");
        
        hud_box.append(&score_lbl);
        hud_box.append(&stake_lbl);
        hud_box.append(&score_them);
    }
    
    // 4. Opponent layer
    let opp_box = window.opponent_box();
    while let Some(child) = opp_box.first_child() {
        opp_box.remove(&child);
    }
    if let Some(players) = &snapshot.players {
        if let Some(opp) = players.get(1) {
            let name_lbl = gtk::Label::new(Some(&opp.name.to_uppercase()));
            name_lbl.add_css_class("opponent-pill");
            opp_box.append(&name_lbl);
            
            let opp_hand = gtk::Box::new(gtk::Orientation::Horizontal, -20);
            let count = opp.hand.as_ref().map(|h| h.len()).unwrap_or(3);
            for _ in 0..count {
                opp_hand.append(&create_card_widget(None));
            }
            opp_box.append(&opp_hand);
        }
    }
    
    // 5. Center Table Layer
    let center_box = window.center_box();
    while let Some(child) = center_box.first_child() {
        center_box.remove(&child);
    }
    if let Some(hand) = &snapshot.current_hand {
        let vira_stack = gtk::Box::new(gtk::Orientation::Vertical, 8);
        vira_stack.append(&gtk::Label::new(Some("VIRA")));
        if let Some(vira) = &hand.vira {
            vira_stack.append(&create_card_widget(Some(vira)));
        }
        center_box.append(&vira_stack);
        
        let played_box = gtk::Box::new(gtk::Orientation::Horizontal, -30);
        played_box.set_size_request(180, 160);
        played_box.set_valign(gtk::Align::Center);
        played_box.set_halign(gtk::Align::Center);
        
        if let Some(rounds) = &hand.round_cards {
            for playing in rounds.iter() {
                played_box.append(&create_card_widget(Some(&playing.card)));
            }
        }
        center_box.append(&played_box);
        
        let manilha_stack = gtk::Box::new(gtk::Orientation::Vertical, 8);
        manilha_stack.append(&gtk::Label::new(Some("MANILHA")));
        if let Some(m) = &hand.manilha {
            let m_lbl = gtk::Label::new(None);
            m_lbl.set_size_request(86, 124);
            m_lbl.set_markup(&format!("<span size='34000'><b>{}</b></span>", m));
            m_lbl.add_css_class("manilha-area");
            m_lbl.add_css_class("manilha-glow");
            manilha_stack.append(&m_lbl);
        }
        center_box.append(&manilha_stack);
    }
    
    // 6. My Hand and Controls Layer
    let bottom_box = window.bottom_box();
    while let Some(child) = bottom_box.first_child() {
        bottom_box.remove(&child);
    }

    // Action buttons
    let action_box = gtk::Box::new(gtk::Orientation::Horizontal, 16);
    action_box.set_halign(gtk::Align::Center);
    
    let btn_truco = gtk::Button::with_label("TRUCO");
    btn_truco.add_css_class("btn-truco");
    let core_truco = core.clone();
    btn_truco.connect_clicked(move |_| {
        let _ = core_truco.dispatch(r#"{"kind":"request_truco"}"#);
    });

    let btn_accept = gtk::Button::with_label("ACEITAR");
    btn_accept.add_css_class("btn-accept");
    let core_accept = core.clone();
    btn_accept.connect_clicked(move |_| {
        let _ = core_accept.dispatch(r#"{"kind":"accept_truco"}"#);
    });

    let btn_refuse = gtk::Button::with_label("CORRER");
    btn_refuse.add_css_class("btn-refuse");
    let core_refuse = core.clone();
    btn_refuse.connect_clicked(move |_| {
        let _ = core_refuse.dispatch(r#"{"kind":"refuse_truco"}"#);
    });

    action_box.append(&btn_truco);
    action_box.append(&btn_accept);
    action_box.append(&btn_refuse);
    bottom_box.append(&action_box);

    if snapshot.turn_player == Some(0) {
        let lbl = gtk::Label::new(Some("SUA VEZ"));
        lbl.add_css_class("turn-pill");
        bottom_box.append(&lbl);
    }

    if let Some(players) = &snapshot.players {
        if let Some(me) = players.get(0) {
            let my_hand = gtk::Box::new(gtk::Orientation::Horizontal, 16);
            if let Some(cards) = &me.hand {
                for c in cards.iter() {
                    let card_widget = create_card_widget(Some(c));
                    card_widget.add_css_class("card-clickable");
                    
                    let gesture = gtk::GestureClick::new();
                    let card_rank = c.rank.clone();
                    let card_suit = c.suit.clone();
                    let core_clone = core.clone();
                    
                    gesture.connect_pressed(move |_, _, _, _| {
                        let play_json = format!(r#"{{"kind":"play_card","payload":{{"card":{{"Rank":"{}","Suit":"{}"}}}}}}"#, card_rank, card_suit);
                        let _ = core_clone.dispatch(&play_json);
                    });
                    
                    card_widget.add_controller(gesture);
                    my_hand.append(&card_widget);
                }
            }
            bottom_box.append(&my_hand);
        }
    }
}

fn create_card_widget(card_opt: Option<&models::Card>) -> gtk::Box {
    let container = gtk::Box::new(gtk::Orientation::Vertical, 0);
    container.set_size_request(86, 124);
    
    if let Some(card) = card_opt {
        container.add_css_class("card");
        if card.is_red() {
            container.add_css_class("card-red");
        }
        
        let top_str = format!("<span size='large'><b>{}</b></span>\n{}", card.rank, card.suit_symbol());
        let top_lbl = gtk::Label::new(None);
        top_lbl.set_markup(&top_str);
        top_lbl.set_halign(gtk::Align::Start);
        top_lbl.set_valign(gtk::Align::Start);
        top_lbl.set_margin_top(8);
        top_lbl.set_margin_start(8);
        
        let center_lbl = gtk::Label::new(None);
        center_lbl.set_markup(&format!("<span size='32000'><b>{}</b></span>", card.suit_symbol()));
        center_lbl.set_vexpand(true);
        
        let bottom_lbl = gtk::Label::new(None);
        bottom_lbl.set_markup(&top_str);
        bottom_lbl.set_halign(gtk::Align::End);
        bottom_lbl.set_valign(gtk::Align::End);
        bottom_lbl.set_margin_bottom(8);
        bottom_lbl.set_margin_end(8);
        
        container.append(&top_lbl);
        container.append(&center_lbl);
        container.append(&bottom_lbl);
    } else {
        container.add_css_class("card-back");
    }
    
    container
}
